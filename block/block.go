package block

import (
	"fmt"
	"time"
)

type Block struct {
	Number    uint64
	Hash      string
	Parent    *Block
	Timestamp time.Time
	Data      string
}

func CreateBlock(parent *Block, data string, blockNumber uint64) *Block {
	return &Block{
		Number:    blockNumber,
		Hash:      fmt.Sprintf("%x", time.Now().UnixNano()),
		Parent:    parent,
		Timestamp: time.Now(),
		Data:      data,
	}
}
