package main

import (
	"deukyunlee/hotstuff/core/network"
	"deukyunlee/hotstuff/logging"
	"flag"
	"os"
)

var (
	logger = logging.GetLogger()
)

var Id int
var runningInDocker bool

func init() {
	idPtr := flag.Int("id", 1, "hotstuff Node ID")

	flag.Parse()

	Id = *idPtr
	runningInDocker = os.Getenv("RUNNING_IN_DOCKER") == "true"
}

func main() {
	node := network.StartNewNode(Id, runningInDocker)
	logger.Info("starting server: %v", node)
	select {}
}
