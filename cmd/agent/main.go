package main

import (
	"github.com/megaded/metrictmr/internal/agent"
)

func main() {
	a := agent.CreateAgent()
	a.StarSend()
}
