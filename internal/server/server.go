package server

import (
	"net/http"

	"github.com/megaded/metrictmr/internal/logger"
	"github.com/megaded/metrictmr/internal/server/handler"
	"github.com/megaded/metrictmr/internal/server/handler/config"
	"github.com/megaded/metrictmr/internal/server/handler/storage"
	"github.com/megaded/metrictmr/internal/server/middleware"
)

type Server struct {
	Handler http.Handler
	Address string
}

func (s *Server) Start() (err error) {
	return http.ListenAndServe(s.Address, s.Handler)
}

type Configer interface {
	GetAddress() string
}

type Listener interface {
	Start() (err error)
}

func CreateServer() (s Listener) {
	server := &Server{}
	logger.SetupLogger("Info")

	//gzipCompressor := &middleware.GZipCompressor{}
	serverConfig := config.GetConfig()
	logger.Log.Info(serverConfig.FilePath)
	storage := storage.NewFileStorage(*serverConfig.StoreInterval, serverConfig.FilePath, *serverConfig.Restore)
	server.Handler = handler.CreateRouter(storage, middleware.Logger, middleware.GzipMiddleware)
	server.Address = serverConfig.Address
	return server
}
