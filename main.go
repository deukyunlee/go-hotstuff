package main

import (
	"deukyunlee/hotstuff/core/network"
	"deukyunlee/hotstuff/logging"
	"flag"
)

var (
	logger = logging.GetLogger()
)

var Id int

func init() {
	idPtr := flag.Int("id", 1, "hotstuff Node ID")

	flag.Parse()

	Id = *idPtr
}

func main() {
	network.StartNewServer(Id)
	logger.Info("starting server")
	select {}
}
