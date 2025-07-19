package server

import (
	"context"
	"net/http"
	"os"

	"github.com/megaded/metrictmr/internal/logger"
	"github.com/megaded/metrictmr/internal/server/handler"
	"github.com/megaded/metrictmr/internal/server/handler/config"
	"github.com/megaded/metrictmr/internal/server/handler/storage"
	"github.com/megaded/metrictmr/internal/server/middleware"
	"go.uber.org/zap"
)

type Server struct {
	Handler   http.Handler
	Address   string
	Cert      string
	PublicKey string
}

func (s *Server) Start(ctx context.Context) {
	server := http.Server{Addr: s.Address, Handler: s.Handler}
	go func() {
		<-ctx.Done()
		server.Shutdown(ctx)
	}()
	if s.Cert != "" && s.PublicKey != "" {
		err := server.ListenAndServeTLS(s.Cert, s.PublicKey)
		if err != nil {
			panic(err)
		}
	} else {
		err := server.ListenAndServe()
		if err != nil {
			panic(err)
		}
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
	cert, publicKey, err := getFilesFromPath(serverConfig.CryptoKey)
	if err == nil {
		server.Cert = cert
		server.PublicKey = publicKey
	}
	storage := storage.CreateStorage(ctx, *serverConfig)
	server.Handler = handler.CreateRouter(storage, middleware.Logger, middleware.GzipMiddleware)
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

func getFilesFromPath(cryptoPath string) (string, string, error) {
	files, err := os.ReadDir(cryptoPath)
	if err != nil {
		return "", "", err
	}

	var cert, key string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Name() == "certificate.pem" {
			cert = cryptoPath + "/certificate.pem"
		}
		if file.Name() == "private.key" {
			key = cryptoPath + "/private.key"
		}
	}

	return cert, key, nil
}
