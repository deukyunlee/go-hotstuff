package network

import "deukyunlee/hotstuff/core/consensus"

type Block struct {
	ParentHash []byte
	Hash       []byte
	height     uint64
	QC         *QuorumCert
	Commited   bool
}

type QuorumCert struct {
	BlockHash []byte
	MsgType   consensus.MsgType
	viewNum   uint64
	signature []byte
}
