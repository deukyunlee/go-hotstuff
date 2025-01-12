package block

import (
	"deukyunlee/hotstuff/types"
	"fmt"
	"time"
)

type Block struct {
	Number    uint64
	Hash      string
	Parent    *Block
	Timestamp time.Time
	Data      string
	View      uint64
	ParentQC  *types.QuorumCertificate
    ProposerID uint64
    Height    uint64	
}

func CreateBlock(parent *Block, data string) *Block {
	var blockNumber uint64 = 1
	if parent != nil {
		blockNumber = parent.Number + 1
	}
	return &Block{
		Number:    blockNumber,
		Hash:      fmt.Sprintf("%x", time.Now().UnixNano()),
		Parent:    parent,
		Timestamp: time.Now(),
		Data:      data,
	}
}
