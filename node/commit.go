package node

import (
	"deukyunlee/hotstuff/message"
)

func (n *Node) ReplyCommit(msg message.Message) {
	block := msg.Block

	logger.Infof("Replying to Commit with vote for block %s\n", block.Hash)

	n.Unicast(message.Message{
		Type:     message.CommitReply,
		Block:    block,
		SenderID: n.ID,
		View:     msg.View,
	}, n.GetLeaderID())
}

func (n *Node) HandleCommitReply(msg message.Message) {
	if !n.IsLeaderNode() {
		logger.Errorf("Cannot handle CommitReply. Only the leader can process this message.\n")
		return
	}

	n.MsgBuffer[message.CommitReply] = append(n.MsgBuffer[message.CommitReply], msg)
	logger.Infof("Leader Node %d: Received CommitReply from Node %d\n", n.ID, msg.SenderID)

	if uint64(len(n.MsgBuffer[message.CommitReply])) >= n.Quorum {
		logger.Infof("Leader Node %d: Quorum reached in Commit phase. Block committed.\n", n.ID)
		n.Committed = append(n.Committed, n.PendingBlock)

		n.Broadcast(message.Message{
			Type:     message.Decide,
			Block:    n.PendingBlock,
			SenderID: n.ID,
			View:     n.View,
		})

		n.PendingBlock = nil
		n.RotateView()
	}
}

func (n *Node) RotateView() {
	n.View++
	logger.Infof("Rotating to View %d. New Leader: Node %d\n", n.View, n.GetLeaderID())
}
