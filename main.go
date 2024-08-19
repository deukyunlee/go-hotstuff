package main

import (
	"deukyunlee/hotstuff/core/network"
	"os"
)

func main() {

	nodeID := os.Args[1]

	server := network.NewServer(nodeID)

	server.Start()

	select {}
}
