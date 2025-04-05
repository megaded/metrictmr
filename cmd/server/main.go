package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/megaded/metrictmr/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()
	s := server.CreateServer(ctx)
	s.Start(ctx)
}
