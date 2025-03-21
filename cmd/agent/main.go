package main

import (
	"github.com/megaded/metrictmr/internal/agent"
	"github.com/megaded/metrictmr/internal/logger"
)

func main() {
	logger.SetupLogger("Info")
	a := agent.CreateAgent()
	a.StartSend()
}
