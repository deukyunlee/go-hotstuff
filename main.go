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

var Id uint64
var runningInDocker bool

func init() {
	idPtr := flag.Uint64("id", 1, "hotstuff Node ID")

	flag.Parse()

	Id = *idPtr
	runningInDocker = os.Getenv("RUNNING_IN_DOCKER") == "true"
}

func main() {
	node := network.StartNewNode(Id, runningInDocker)
	logger.Info("starting server: %v", node)

	// INITIAL LEADER is node 1
	logger.Infof("current state: %v, Leader: %v", node.CurrentState, node.CurrentState.GetLeader())

	node.SendInitialRequestMsg()
	select {}
}
