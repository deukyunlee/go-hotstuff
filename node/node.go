package node

import (
	"deukyunlee/hotstuff/block"
	"deukyunlee/hotstuff/logging"
	"deukyunlee/hotstuff/message"
	"encoding/json"
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
	}
}

func StartNewNode(nodeId uint64) *Node {
	return setNode(nodeId, NodeTable[nodeId])
}

func setNode(nodeId uint64, address string) *Node {

	node := NewNode(nodeId, 4, 3)

	go node.startServer(nodeId, address)

	for otherId, otherAddress := range NodeTable {
		if otherId != nodeId {
			err := node.startClientWithRetry(nodeId, otherId, otherAddress)
			if err != nil {
				logger.Errorf("Failed to connect to node %d after %d attempts: %v\n", otherId, RetryCount, err)
				panic(err)
			}
		}
	}

	return node
}

func (n *Node) startClientWithRetry(nodeId uint64, otherId uint64, otherAddress string) error {
	var err error
	var conn net.Conn
	for i := uint64(0); i < RetryCount; i++ {
		conn, err = n.startClient(nodeId, otherId, otherAddress)

		if err == nil {
			logger.Infof("Node %d: Connected to Node %d [LOCAL: %s] [REMOTE: %s]\n", nodeId, otherId, conn.LocalAddr(), conn.RemoteAddr())
			n.Connections[i] = conn
			break
		}

		logger.Errorf("Failed to connect to node %d (attempt %d/%d): %v\n", otherId, i+1, RetryCount, err)

		time.Sleep(ConnectionDelay)
	}
	return err
}

func (n *Node) startClient(id, otherID uint64, address string) (net.Conn, error) {
	conn, err := net.Dial(TcpNetworkType, address)
	if err != nil {
		logger.Errorf("Node %d: Error connecting to Node %d at %s: %v\n", id, otherID, address, err)
		return nil, err
	}

	logger.Infof("Node %d: Connected to Node %d [LOCAL: %s] [REMOTE: %s]\n", id, otherID, conn.LocalAddr(), conn.RemoteAddr())
	return conn, nil
}

func (n *Node) startServer(nodeId uint64, address string) {
	ln, err := net.Listen(TcpNetworkType, address)
	if err != nil {
		logger.Errorf("Node %d: Error starting server: %v\n", nodeId, err)
	}

	if NodeTable[nodeId] == "" {
		panic("Unable to get server info")
	}

	for {
		logger.Infof("Waiting for connection on %s", address)

		conn, err := n.startServerWithRetry(nodeId, ln)

		if err != nil {
			logger.Errorf("Failed to connect to node %d after %d attempts: %d\n", nodeId, RetryCount, err)
			panic(err)
		}

		logger.Infof("[%d]Connection accepted from %s\n", n.ID, conn.RemoteAddr().String())

		n.LocalConnection = conn

		go n.handleConnection(nodeId, conn)
	}
}

func (n *Node) handleConnection(id uint64, conn net.Conn) {
	logger.Infof("Node %d: Accepted connection from %s\n", id, conn.RemoteAddr().String())

	buffer := make([]byte, 1024)

	rawMsg, err := conn.Read(buffer)

	var msg message.Message
	err = json.Unmarshal(buffer[:rawMsg], &msg)
	if err != nil {
		logger.Errorf("Node %d: Error unmarshalling message: %v", id, err)
		return
	}

	logger.Infof("[NodeId: %d] rawMsg: %v", n.ID, msg)

	go n.ReceiveMessage(msg)
}

func (n *Node) startServerWithRetry(nodeId uint64, ln net.Listener) (conn net.Conn, err error) {
	for i := 0; i < RetryCount; i++ {
		conn, err = ln.Accept()
		if err == nil {
			logger.Infof("Node %d: accepting connection: %v\n", nodeId, err)

			break
		}

		logger.Errorf("Failed to connect to node %d (attempt %d/%d): %d\n", nodeId, i+1, RetryCount, err)

		time.Sleep(ConnectionDelay)
	}
	return conn, err
}

func (n *Node) ReceiveMessage(msg message.Message) {
	n.Mutex.Lock()
	defer n.Mutex.Unlock()

	logger.Infof("Node %d: Received message %v\n", n.ID, msg)
	n.MsgBuffer[msg.Type] = append(n.MsgBuffer[msg.Type], msg)

	switch msg.Type {
	case message.Prepare:
		if !n.IsLeaderNode() {
			n.ReplyPrepare(msg.Block)
		}
	case message.PrepareReply:
		if n.IsLeaderNode() {
			n.HandlePrepareReply(msg)
		}
	case message.PreCommit:
		if !n.IsLeaderNode() {
			n.ReplyPreCommit(msg.Block)
		}
	case message.PreCommitReply:
		if n.IsLeaderNode() {
			n.HandlePreCommitReply(msg)
		}
	case message.Commit:
		if !n.IsLeaderNode() {
			n.ReplyCommit(msg.Block)
		}
	case message.CommitReply:
		if n.IsLeaderNode() {
			n.HandleCommitReply(msg)
		}
	default:
		panic("unhandled default case")
	}
}
