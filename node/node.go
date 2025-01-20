package node

import (
	"deukyunlee/hotstuff/block"
	"deukyunlee/hotstuff/logging"
	"deukyunlee/hotstuff/message"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var logger = logging.GetLogger()

var NodeTable = map[uint64]string{
	1: "localhost:1111",
	2: "localhost:1112",
	3: "localhost:1113",
	4: "localhost:1114",
}

const (
	TcpNetworkType  string = "tcp"
	RetryCount             = 5
	ConnectionDelay        = 2 * time.Second
)

type Node struct {
	ID              uint64
	View            uint64
	TotalNodes      uint64
	BlockChain      map[string]*block.Block
	MsgBuffer       map[message.MessageType][]message.Message
	Quorum          uint64
	Mutex           sync.Mutex
	Committed       []*block.Block
	PendingBlock    *block.Block
	Connections     map[uint64]net.Conn
	LocalConnection net.Conn
	IsLeader        bool
	RecentIntervals []time.Duration 
	WindowSize      int          
	TotalInterval   time.Duration     
	BlockCount      int              
	LastBlockTime   time.Time          
}

func NewNode(id, totalNodes, quorum uint64) *Node {

	return &Node{
		ID:          id,
		View:        0,
		TotalNodes:  totalNodes,
		BlockChain:  make(map[string]*block.Block),
		MsgBuffer:   make(map[message.MessageType][]message.Message),
		Connections: make(map[uint64]net.Conn),
		Quorum:      quorum,
		RecentIntervals: make([]time.Duration, 0),
		WindowSize: 10,  
		TotalInterval: 0,
		BlockCount: 0,
		LastBlockTime: time.Now(),
	}
}

func StartNewNode(nodeId uint64) *Node {
	node := NewNode(nodeId, 4, 3)
	
	serverReady := make(chan bool)
	go node.startServer(nodeId, NodeTable[nodeId], serverReady)
	
	<-serverReady
	logger.Infof("Node %d: Server is ready and listening", nodeId)
	
	time.Sleep(time.Second)
	
	for otherId, otherAddress := range NodeTable {
		logger.Infof("Connecting to Node %d\n", otherId)
		if otherId != nodeId {
			err := node.startClientWithRetry(otherId, otherAddress)
			if err != nil {
				logger.Errorf("Node %d: Failed to connect to node %d after %d attempts: %v\n", 
					nodeId, otherId, RetryCount, err)
				panic(err)
			}
		}
	}
	
	return node
}

func (n *Node) startServer(nodeId uint64, address string, ready chan bool) {
	ln, err := net.Listen(TcpNetworkType, address)
	if err != nil {
		logger.Errorf("Node %d: Error starting server: %v\n", nodeId, err)
		panic(err)
	}
	
	logger.Infof("Node %d: Server started on %s", nodeId, address)
	ready <- true
	
	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.Errorf("Node %d: Failed to accept connection: %v\n", nodeId, err)
			continue
		}
		
		logger.Infof("Node %d: Connection accepted from %s\n", nodeId, conn.RemoteAddr().String())
		n.LocalConnection = conn
		go n.handleConnection(conn)
	}
}

func (n *Node) startClientWithRetry(otherId uint64, otherAddress string) error {
	var err error
	var conn net.Conn
	
	const maxRetries = 15
	const initialDelay = 1 * time.Second
	
	for i := 0; i < maxRetries; i++ {
		conn, err = n.startClient(otherId, otherAddress)
		if err == nil {
			n.Connections[otherId] = conn
			return nil
		}
		
		delay := initialDelay * time.Duration(i+1)
		logger.Infof("Retrying connection to Node %d in %v (attempt %d/%d)",otherId, delay, i+1, maxRetries)
		time.Sleep(delay)
	}
	
	return fmt.Errorf("failed to connect after %d attempts: %v", maxRetries, err)
}

func (n *Node) startClient(otherID uint64, address string) (net.Conn, error) {
	conn, err := net.Dial(TcpNetworkType, address)
	if err != nil {
		logger.Errorf("Error connecting to Node %d at %s: %v\n", otherID, address, err)
		return nil, err
	}

	logger.Infof("Connected to Node %d [LOCAL: %s] [REMOTE: %s]\n", otherID, conn.LocalAddr(), conn.RemoteAddr())
	return conn, nil
}

func (n *Node) startServerWithRetry(nodeId uint64, ln net.Listener) (conn net.Conn, err error) {
	for i := 0; i < RetryCount; i++ {
		conn, err = ln.Accept()
		if err == nil {
			logger.Infof("accepting connection: %v\n", conn.RemoteAddr())

			break
		}

		logger.Errorf("Failed to connect to node %d (attempt %d/%d): %d\n", nodeId, i+1, RetryCount, err)

		time.Sleep(ConnectionDelay)
	}
	return conn, err
}

func (n *Node) handleConnection(conn net.Conn) {
	logger.Infof("Accepted connection from %s\n", conn.RemoteAddr().String())
	
	for {
		// 먼저 메시지 길이를 읽음
		lengthBuf := make([]byte, 4)
		_, err := io.ReadFull(conn, lengthBuf)
		if err != nil {
			if err != io.EOF {
				logger.Errorf("Error reading message length: %v", err)
			}
			return
		}
		
		// 메시지 길이 해석
		messageLength := binary.BigEndian.Uint32(lengthBuf)
		
		// 메시지 본문 읽기
		messageBuf := make([]byte, messageLength)
		_, err = io.ReadFull(conn, messageBuf)
		if err != nil {
			logger.Errorf("Error reading message body: %v", err)
			return
		}
		
		var msg message.Message
		err = json.Unmarshal(messageBuf, &msg)
		if err != nil {
			logger.Errorf("Error unmarshalling message: %v\nRaw message: %s", err, string(messageBuf))
			continue
		}
		
		go n.ReceiveMessage(msg)
	}
}

func (n *Node) ReceiveMessage(msg message.Message) {
	n.Mutex.Lock()
	defer n.Mutex.Unlock()

	logger.Infof("Received message %v\n", msg)
	n.MsgBuffer[msg.Type] = append(n.MsgBuffer[msg.Type], msg)

	switch msg.Type {
	case message.Prepare:
		if !n.IsLeaderNode() {
			n.ReplyPrepare(msg)
		}
	case message.PrepareReply:
		if n.IsLeaderNode() {
			n.HandlePrepareReply(msg)
		}
	case message.PreCommit:
		if !n.IsLeaderNode() {
			n.ReplyPreCommit(msg)
		}
	case message.PreCommitReply:
		if n.IsLeaderNode() {
			n.HandlePreCommitReply(msg)
		}
	case message.Commit:
		if !n.IsLeaderNode() {
			n.ReplyCommit(msg)
		}
	case message.CommitReply:
		if n.IsLeaderNode() {
			n.HandleCommitReply(msg)
		}
	case message.Decide:
		n.HandleDecide(msg)

	default:
		panic("unhandled default case")
	}
}

func (n *Node) ValidateBlock(b *block.Block) error {
	if b == nil {
		return errors.New("block is nil")
	}

	if b.Hash == "" {
		return errors.New("block hash is empty")
	}

	if b.Parent == nil {
		return errors.New("previous block hash is empty")
	}

	//if b.View < n.View {
	//	return fmt.Errorf("block view %d is less than current node view %d", b.View, n.View)
	//}

	//if !n.verifySignature(b.Signature, b.Hash, b.ProposerID) {
	//	return errors.New("invalid block signature")
	//}

	return nil
}