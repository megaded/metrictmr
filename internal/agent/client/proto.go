package client

import (
	"context"
	"time"

	"github.com/megaded/metrictmr/internal/agent/collector"
	"github.com/megaded/metrictmr/internal/agent/config"
	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/logger"
	pb "github.com/megaded/metrictmr/internal/proto"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProtoClient struct {
	Config Configer
	Client pb.MetricServiceClient
	hostIp string
}

func CreateProtoClient() *ProtoClient {
	pClient := ProtoClient{}
	pClient.Config = config.GetConfig()
	conn, err := grpc.NewClient(pClient.Config.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
	defer conn.Close()
	ip, err := getLocalIP()
	if err != nil {
		pClient.hostIp = ip.String()
	}
	pClient.Client = pb.NewMetricServiceClient(conn)
	return &pClient
}

func (p *ProtoClient) StartSend(ctx context.Context) {
	rateLimit := p.Config.GetRateLimit()
	mch := make(chan collector.Metric, rateLimit)
	metricCollector := &collector.MetricCollector{}
	addr := p.Config.GetAddress()
	pollInterval := p.Config.GetPoolInterval()
	group, ctxCancel := errgroup.WithContext(ctx)

	for w := 0; w <= rateLimit; w++ {
		group.Go(func() error {
			return protoWorker(ctxCancel, addr, p.hostIp, p.Client, mch)
		})

	}
	group.Go(func() error {
		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
		defer ticker.Stop()
		defer close(mch)
		for {
			select {
			case <-ctxCancel.Done():
				return ctxCancel.Err()
			case <-ticker.C:
				m := metricCollector.GetRunTimeMetrics()
				select {
				case mch <- m:
				case <-ctxCancel.Done():
					return ctxCancel.Err()
				}

			}
		}
	})

	if err := group.Wait(); err != nil {
		logger.Log.Error("Agent error", zap.Error(err))
	}
}

func protoWorker(ctx context.Context, addr string, ip string, client pb.MetricServiceClient, jobs <-chan collector.Metric) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case m, ok := <-jobs:
			if !ok {
				return nil
			}
			if err := sendMetric(ctx, m, addr, client, ip); err != nil {
				logger.Log.Warn("send metric error", zap.Error(err))
			}
		}
	}
}

func sendMetric(ctx context.Context, c collector.Metric, addr string, client pb.MetricServiceClient, ip string) error {
	if len(c.GaugeMetrics) == 0 && len(c.CounterMetrics) == 0 {
		logger.Log.Info("Отправка метрик. Метрик нет")
		return nil
	}
	d := make([]*pb.Metric, 0, len(c.GaugeMetrics)+len(c.CounterMetrics))
	for _, v := range c.GaugeMetrics {
		value := float32(v.Value)
		m := pb.Metric{Id: string(v.Name), Type: data.MTypeGauge, Value: value}
		d = append(d, &m)
	}
	for _, v := range c.CounterMetrics {
		m := pb.Metric{Id: string(v.Name), Type: data.MTypeCounter, Delta: v.Value}
		d = append(d, &m)
	}
	req := pb.UpdateMetricsRequest{}

	req.Metrics = d
	_, err := client.Update(ctx, &req)
	return err
}
