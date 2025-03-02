package server

import (
	"net/http"

	"github.com/megaded/metrictmr/internal/server/handler"
	"github.com/megaded/metrictmr/internal/server/handler/config"
	"github.com/megaded/metrictmr/internal/server/handler/storage"
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
	storage := storage.NewStorage()
	server.Handler = handler.CreateRouter(storage)
	server.config = config.GetConfig()
	return server
}
