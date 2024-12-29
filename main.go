package main

import (
	"deukyunlee/hotstuff/block"
	"deukyunlee/hotstuff/node"
	"flag"
)

var Id uint64

func init() {
	idPtr := flag.Uint64("id", 1, "hotstuff Node ID")

	flag.Parse()

	Id = *idPtr
}

func main() {
	n := node.StartNewNode(Id)

	if n.IsLeaderNode() {
		genesisBlock := block.CreateBlock(nil, "Genesis Block", 0)

		n.BlockChain[genesisBlock.Hash] = genesisBlock

		newBlock := block.CreateBlock(genesisBlock, "Block 1", genesisBlock.Number+1)

		n.Propose(newBlock)
	}
}
