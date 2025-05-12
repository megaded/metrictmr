package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/megaded/metrictmr/internal/agent"
	"github.com/megaded/metrictmr/internal/logger"
)

func main() {
	logger.SetupLogger("Info")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()
	a := agent.CreateAgent()
	a.StartSend(ctx)
}
