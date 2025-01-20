package main

import (
	"deukyunlee/hotstuff/block"
	"deukyunlee/hotstuff/message"
	"deukyunlee/hotstuff/node"
	"flag"
	"time"
)

var Id uint64
var GenerateSecrets bool

func init() {
	idPtr := flag.Uint64("id", 1, "hotstuff Node ID")
	genFlag := flag.Bool("genSecrets", false, "Generate partial secrets and save to files")

	flag.Parse()

	Id = *idPtr
	GenerateSecrets = *genFlag
}

// node.ValidateBlock 변경 로직: 블록 validation 로직 추가: node.ValidateBlock
// message에 view 추가: message.Message
// decide phase 추가: node.decide

// TODO: 자신이 보고 있는 view와 message의 view가 일치하는지 확인
// TODO: 밸리데이터 sig 검증 추가

const (
	cnt = 4 // 총 노드 수
	t   = 3 // threshold
)

func main() {
	if GenerateSecrets {
		node.GenerateAndSaveSecrets(cnt, t)

		return
	}

	n := node.StartNewNode(Id)

	// view 증가 시점
	// 블록이 정상적으로 Commit(또는 Decide)된 뒤 다음 라운드로 넘어갈 때
	// 현재 라운드의 리더가 장애가 나서 타임아웃(ViewChange)가 발생할 때
	
	if n.IsLeaderNode() {
		genesisBlock := block.CreateBlock(nil, "Genesis Block")

		n.BlockChain[genesisBlock.Hash] = genesisBlock

		newBlock := block.CreateBlock(genesisBlock, "Block 1")

		msg := message.Message{
			Type:     message.Prepare,
			Block:    newBlock,
			SenderID: n.ID,
			View:     0,
		}

		for {
			time.Sleep(100 * time.Millisecond)

			if (len(n.Connections)>=3){
				n.Propose(msg)
				break
			}
		}
	}

	for {
		time.Sleep(100 * time.Millisecond)
	}
}
