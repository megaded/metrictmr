package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/megaded/metrictmr/internal/server"
)

func main() {
	for i, v := range os.Args[1:] {
		fmt.Println(i+1, v)
	}
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
