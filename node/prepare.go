package node

import (
	"deukyunlee/hotstuff/block"
	"deukyunlee/hotstuff/message"
)

func (n *Node) ReplyPrepare(block *block.Block) {
	logger.Infof("Node %d: Replying to Prepare with vote for block %s\n", n.ID, block.Hash)
	n.Unicast(message.Message{
		Type:     message.PrepareReply,
		Block:    block,
		SenderID: n.ID,
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
		})
	}
}
