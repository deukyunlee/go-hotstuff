package node

import (
	"deukyunlee/hotstuff/message"
)

func (n *Node) ReplyPrepare(msg message.Message) {
	block := msg.Block

	logger.Infof("Replying to Prepare with vote for block %s\n", block.Hash)

	if err := n.ValidateBlock(block); err != nil {
		logger.Warnf("Invalid block. Err: %s\n", err)
		return
	}

	n.Unicast(message.Message{
		Type:     message.PrepareReply,
		Block:    block,
		SenderID: n.ID,
		View:     msg.View,
	}, n.GetLeaderID())
}

func (n *Node) HandlePrepareReply(msg message.Message) {
	if !n.IsLeaderNode() {
		logger.Errorf("Node %d: Cannot handle PrepareReply. Only the leader can process this message.\n", n.ID)
		return
	}

	n.MsgBuffer[message.PrepareReply] = append(n.MsgBuffer[message.PrepareReply], msg)
	logger.Infof("Leader Node %d: Received PrepareReply from Node %d\n", n.ID, msg.SenderID)

	if uint64(len(n.MsgBuffer[message.PrepareReply])) >= n.Quorum {
		logger.Infof("Leader Node %d: Quorum reached in Prepare phase.\n", n.ID)
		n.Broadcast(message.Message{
			Type:     message.PreCommit,
			Block:    n.PendingBlock,
			SenderID: n.ID,
			View:     n.View,
		})
	}
}
