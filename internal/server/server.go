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
	config  Configer
}

func (s *Server) Start() (err error) {
	return http.ListenAndServe(s.config.GetAddress(), s.Handler)
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
	storage := storage.NewStorage()
	server.Handler = handler.CreateRouter(storage, middleware.Logger)
	server.config = config.GetConfig()
	return server
}
