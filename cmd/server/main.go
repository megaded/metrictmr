package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/megaded/metrictmr/internal/logger"
	"github.com/megaded/metrictmr/internal/server"
	"go.uber.org/zap"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-sigChan
		cancel()
	}()
	logger.Log.Info("Build information",
		zap.String("version", buildVersion),
		zap.String("date", buildDate),
		zap.String("commit", buildCommit),
	)
	s := server.CreateServer(ctx)
	s.Start(ctx)
}
