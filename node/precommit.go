package node

import (
	"deukyunlee/hotstuff/block"
	"deukyunlee/hotstuff/message"
)

func (n *Node) ReplyPreCommit(block *block.Block) {
	logger.Infof("Node %d: Replying to PreCommit with vote for block %s\n", n.ID, block.Hash)
	n.Unicast(message.Message{
		Type:     message.PreCommitReply,
		Block:    block,
		SenderID: n.ID,
	}, n.GetLeaderID())
}

func (n *Node) HandlePreCommitReply(msg message.Message) {
	if !n.IsLeaderNode() {
		logger.Errorf("Node %d: Cannot handle PreCommitReply. Only the leader can process this message.\n", n.ID)
		return
	}

	n.MsgBuffer[message.PreCommitReply] = append(n.MsgBuffer[message.PreCommitReply], msg)
	logger.Infof("Leader Node %d: Received PreCommitReply from Node %d\n", n.ID, msg.SenderID)

	if uint64(len(n.MsgBuffer[message.PreCommitReply])) >= n.Quorum {
		logger.Infof("Leader Node %d: Quorum reached in PreCommit phase.\n", n.ID)
		n.Broadcast(message.Message{
			Type:     message.Commit,
			Block:    n.PendingBlock,
			SenderID: n.ID,
		})
	}
}
