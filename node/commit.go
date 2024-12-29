package node

import (
	"deukyunlee/hotstuff/block"
	"deukyunlee/hotstuff/message"
)

func (n *Node) ReplyCommit(block *block.Block) {
	logger.Infof("Node %d: Replying to Commit with vote for block %s\n", n.ID, block.Hash)
	n.Unicast(message.Message{
		Type:     message.CommitReply,
		Block:    block,
		SenderID: n.ID,
	}, n.GetLeaderID())
}

func (n *Node) HandleCommitReply(msg message.Message) {
	if !n.IsLeaderNode() {
		logger.Errorf("Node %d: Cannot handle CommitReply. Only the leader can process this message.\n", n.ID)
		return
	}

	n.MsgBuffer[message.CommitReply] = append(n.MsgBuffer[message.CommitReply], msg)
	logger.Infof("Leader Node %d: Received CommitReply from Node %d\n", n.ID, msg.SenderID)

	if uint64(len(n.MsgBuffer[message.CommitReply])) >= n.Quorum {
		logger.Infof("Leader Node %d: Quorum reached in Commit phase. Block committed.\n", n.ID)
		n.Committed = append(n.Committed, n.PendingBlock)
		n.PendingBlock = nil
		n.RotateView()
	}
}

func (n *Node) RotateView() {
	n.View++
	logger.Infof("Node %d: Rotating to View %d. New Leader: Node %d\n", n.ID, n.View, n.GetLeaderID())
}
