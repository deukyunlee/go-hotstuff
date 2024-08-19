package core

import (
	"net/rpc"
	"sync"
	"time"
)

type Block struct {
	ID        string
	ParentID  string
	Timestamp time.Time
	Data      []byte
	Signature []byte
}

type MessageType int

const (
	Prepare MessageType = iota
	PreCommit
	Commit
	Decide
)

type Message struct {
	Type      MessageType
	Block     Block
	SenderID  string
	Signature []byte
}

type Response struct {
	Success bool
}

type State struct {
	CurrentBlock Block
	Peers        map[string]string
	Connections  map[string]*rpc.Client
	Mutex        sync.Mutex
}
