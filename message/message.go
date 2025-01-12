package message

import "deukyunlee/hotstuff/block"

type MessageType int

const (
	Req MessageType = iota
	Prepare
	PrepareReply
	PreCommit
	PreCommitReply
	Commit
	CommitReply
	Decide
)

type Message struct {
	Type      MessageType
	Block     *block.Block
	SenderID  uint64
	Signature string
	View      uint64	
    QC        *QuorumCertificate
}
