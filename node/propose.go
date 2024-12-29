package node

import (
	"deukyunlee/hotstuff/block"
	"deukyunlee/hotstuff/message"
)

func (n *Node) Propose(block *block.Block) {
	if !n.IsLeaderNode() {
		logger.Errorf("Node %d: Cannot propose block. Only the leader can propose.\n", n.ID)
		return
	}

	logger.Infof("Leader Node %d (View %d): Proposing block %s\n", n.ID, n.View, block.Hash)
	n.PendingBlock = block
	n.Broadcast(message.Message{
		Type:     message.Prepare,
		Block:    block,
		SenderID: n.ID,
	})
}
