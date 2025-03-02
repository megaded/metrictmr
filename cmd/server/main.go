package main

import (
	"github.com/megaded/metrictmr/internal/server"
)

func main() {
	s := server.CreateServer()
	err := s.Start()
	if err != nil {
		panic(err)
	}

}
