package server

import (
	"context"
	"net/http"

	"github.com/megaded/metrictmr/internal/logger"
	"github.com/megaded/metrictmr/internal/server/handler"
	"github.com/megaded/metrictmr/internal/server/handler/config"
	"github.com/megaded/metrictmr/internal/server/handler/storage"
	"github.com/megaded/metrictmr/internal/server/middleware"
	"go.uber.org/zap"
)

type Server struct {
	Handler http.Handler
	Address string
}

func (s *Server) Start(ctx context.Context) {
	server := http.Server{Addr: s.Address, Handler: s.Handler}
	go func() {
		<-ctx.Done()
		server.Shutdown(ctx)
	}()
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

type Configer interface {
	GetAddress() string
}

type Listener interface {
	Start(ctx context.Context)
}

func CreateServer(ctx context.Context) (s Listener) {
	server := &Server{}
	logger.SetupLogger("Info")
	serverConfig := config.GetConfig()
	logConfig(*serverConfig)
	storage := storage.CreateStorage(ctx, *serverConfig)
	server.Handler = handler.CreateRouter(storage, middleware.Logger, middleware.GzipMiddleware, middleware.Hash(serverConfig.Key))
	server.Address = serverConfig.Address
	return server
}

func logConfig(c config.Config) {
	nConfig := "Config"
	logger.Log.Info(nConfig, zap.String("add", c.Address))
	logger.Log.Info(nConfig, zap.String("path", c.FilePath))
	logger.Log.Info(nConfig, zap.Bool("restore", *c.Restore))
	logger.Log.Info(nConfig, zap.Int("internal", *c.StoreInterval))
	logger.Log.Info(nConfig, zap.String("db conn string", c.DBConnString))
	logger.Log.Info(nConfig, zap.String("key", c.Key))
}
