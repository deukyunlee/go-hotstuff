package main

import (
	"deukyunlee/hotstuff/core/consensus"
	"deukyunlee/hotstuff/core/network"
	"deukyunlee/hotstuff/logging"
	"flag"
	"time"
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
	node := network.StartNewNode(Id)
	logger.Info("starting server: %v", node)
	currentTime := time.Now().UnixNano()

	err := node.Broadcast(consensus.RequestMsg{Timestamp: currentTime, ClientID: 1, Operation: "test", SequenceID: 0})
	for id, er := range err {
		logger.Errorf("[%d] error: %s", id, er.Error())
	}
	time.Sleep(1 * time.Second)

	select {}
}
