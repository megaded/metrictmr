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
		c := make(chan os.Signal, 1) //
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()
	a := agent.CreateAgent()
	a.StartSend(ctx)
}
