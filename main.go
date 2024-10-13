package main

import (
	"deukyunlee/hotstuff/core/network"
	"deukyunlee/hotstuff/logging"
)

var (
	logger = logging.GetLogger()
)

func main() {
	network.StartNewServer()
	logger.Info("starting server")
	select {}
}
