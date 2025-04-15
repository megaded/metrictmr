package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/megaded/metrictmr/internal/agent"
	"github.com/megaded/metrictmr/internal/logger"
)

func main() {
	for i, v := range os.Args[1:] {
		fmt.Println(i+1, v)
	}
	logger.SetupLogger("Info")
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	a := agent.CreateAgent()
	a.StartSend(ctx)
}
