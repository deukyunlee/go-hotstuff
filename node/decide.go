package node

import (
	"deukyunlee/hotstuff/block"
	"deukyunlee/hotstuff/message"
)

func (n *Node) HandleDecide(msg message.Message) {
	logger.Infof("Received Decide for block %s. Finalizing...\n",msg.Block.Hash)

	logger.Infof("Block %d, hash: %s is now decided. Current view: %d\n",msg.Block.Number, msg.Block.Hash, n.View)

	if msg.View >= n.View {
		n.View = msg.View + 1
	}

	newBlock := block.CreateBlock(msg.Block, "")
	newMsg := message.Message{
		Type:     message.Prepare,
		Block:    newBlock,
		SenderID: n.ID,
		View:     n.View,
	}

	n.Committed = append(n.Committed, msg.Block)

	if n.IsLeaderNode() {
		n.Propose(newMsg)
	}
}
