package node

import (
	"deukyunlee/hotstuff/message"
)

func (n *Node) Propose(msg message.Message) {
	block := msg.Block
	if !n.IsLeaderNode() {
		logger.Errorf("Cannot propose block. Only the leader can propose.\n")
		return
	}

	logger.Infof("Leader Node %d (View %d): Proposing block %s\n", n.ID, n.View, block.Hash)
	n.PendingBlock = block
	n.Broadcast(msg)
}
