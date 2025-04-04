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
	defer cancel()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		cancel()
	}()
	s := server.CreateServer(ctx)
	err := s.Start()
	if err != nil {
		panic(err)
	}

}
