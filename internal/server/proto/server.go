package proto

import (
	"context"
	"net"

	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/logger"
	pb "github.com/megaded/metrictmr/internal/proto"
	"github.com/megaded/metrictmr/internal/server"
	"github.com/megaded/metrictmr/internal/server/handler/config"
	"github.com/megaded/metrictmr/internal/server/handler/storage"
	"github.com/megaded/metrictmr/internal/server/proto/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricServer struct {
	pb.UnimplementedMetricServiceServer
	address       string
	trustedSubNet *net.IPNet
	storage       storage.Storager
}

func (s *MetricServer) Update(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if req != nil {
		if len(req.Metrics) > 0 {
			metrics := make([]data.Metric, 0, len(req.Metrics))
			for _, m := range req.Metrics {
				value := float64(m.Value)
				metrics = append(metrics, data.Metric{MType: m.Type, ID: m.Id, Delta: &m.Delta, Value: &value})
			}
			if err := s.storage.Store(ctx, metrics...); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
	}
	resp := new(pb.UpdateMetricsResponse)
	return resp, nil
}

func (s *MetricServer) Start(ctx context.Context) {
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		panic(err)
	}
	var server *grpc.Server
	if s.trustedSubNet != nil {
		interceptor := interceptor.TrustedSubNetInterceptor(*s.trustedSubNet)
		server = grpc.NewServer(grpc.UnaryInterceptor(interceptor))
	} else {
		server = grpc.NewServer()
	}

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()
	pb.RegisterMetricServiceServer(server, &MetricServer{})
	if err := server.Serve(listen); err != nil {
		panic(err)
	}
}

func CreateServer(ctx context.Context) (s server.Listener) {
	server := &MetricServer{}
	logger.SetupLogger("Info")
	serverConfig := config.GetConfig()
	server.address = serverConfig.Address
	if serverConfig.TrustedSubnet != "" {
		_, net, err := net.ParseCIDR(serverConfig.TrustedSubnet)
		if err == nil {
			server.trustedSubNet = net
		}
	}
	server.storage = storage.CreateStorage(ctx, *serverConfig)
	return server
}
