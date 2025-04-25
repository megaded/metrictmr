package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/megaded/metrictmr/internal/agent"
	"github.com/megaded/metrictmr/internal/logger"
)

func main() {
	logger.SetupLogger("Info")
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	a := agent.CreateAgent()
	a.StartSend(ctx)
}
