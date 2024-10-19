package network

import (
	"deukyunlee/hotstuff/core/consensus"
	"deukyunlee/hotstuff/logging"
	"encoding/json"
	"net"
	"time"
)

const (
	TcpNetworkType  string = "tcp"
	RetryCount             = 5
	ConnectionDelay        = 2 * time.Second
)

var (
	logger    = logging.GetLogger()
	NodeTable = map[int]string{
		1: "localhost:1111",
		2: "localhost:1112",
		3: "localhost:1113",
		4: "localhost:1114",
	}
)

func StartNewNode(nodeId int) *Node {
	return setNode(nodeId, NodeTable[nodeId])
}

// setNode starts both server and client and set data for a given node
func setNode(nodeId int, address string) *Node {

	node := NewNode(nodeId)

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

func (node *Node) startClientWithRetry(nodeId int, otherId int, otherAddress string) error {
	var err error
	var conn net.Conn
	for i := 0; i < RetryCount; i++ {
		conn, err = node.startClient(nodeId, otherId, otherAddress)

		if err == nil {
			logger.Infof("Node %d: Connected to Node %d [LOCAL: %s] [REMOTE: %s]\n", nodeId, otherId, conn.LocalAddr(), conn.RemoteAddr())
			node.Connections = append(node.Connections, conn)
			break
		}

		logger.Errorf("Failed to connect to node %d (attempt %d/%d): %v\n", otherId, i+1, RetryCount, err)

		time.Sleep(ConnectionDelay)
	}
	return err
}

// startServer starts the TCP server for the node
func (node *Node) startServer(nodeId int, address string) {
	ln, err := net.Listen(TcpNetworkType, address)
	if err != nil {
		logger.Errorf("Node %d: Error starting server: %v\n", nodeId, err)
	}

	//defer ln.Close()
	if NodeTable[nodeId] == "" {
		panic("Unable to get server info")
	}

	for {
		logger.Infof("Waiting for connection on %s", address)

		conn, err := node.startServerWithRetry(nodeId, ln)

		if err != nil {
			logger.Errorf("Failed to connect to node %d after %d attempts: %d\n", nodeId, RetryCount, err)
			panic(err)
		}

		logger.Infof("Connection accepted from %s", conn.RemoteAddr().String())

		go node.handleConnection(nodeId, conn)
	}
}

func (node *Node) startServerWithRetry(nodeId int, ln net.Listener) (conn net.Conn, err error) {
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

// handleConnection handles incoming connections to the server
func (node *Node) handleConnection(id int, conn net.Conn) {
	logger.Infof("Node %d: Accepted connection from %s\n", id, conn.RemoteAddr().String())

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)

	var rawMsg map[string]interface{}
	err = json.Unmarshal(buffer[:n], &rawMsg)
	if err != nil {
		logger.Errorf("Node %d: Error unmarshalling message: %v", id, err)
		return
	}

	if rawMsgType, ok := rawMsg["msgType"].(float64); ok {

		consensusMsgType := consensus.MsgType(int(rawMsgType))
		switch consensusMsgType {
		case consensus.Request:
			var reqMsg consensus.RequestMsg
			err := json.Unmarshal(buffer[:n], &reqMsg)
			if err != nil {
				logger.Errorf("Node %d: Error unmarshalling message: %v", id, err)
			}
			node.MsgEntrance <- &reqMsg
		case consensus.Prepare:
			var prepareMsg consensus.PrepareMsg
			err := json.Unmarshal(buffer[:n], &prepareMsg)
			if err != nil {
				logger.Errorf("Node %d: Error unmarshalling message: %v", id, err)
			}
			node.MsgEntrance <- &prepareMsg
		case consensus.PreCommit, consensus.Commit, consensus.Decide:
			var consensusMsg consensus.ConsensusMsg
			err := json.Unmarshal(buffer[:n], &consensusMsg)
			if err != nil {
				logger.Errorf("Node %d: Error unmarshalling message: %v", id, err)
			}
			node.MsgEntrance <- &consensusMsg
		default:
			logger.Errorf("Node %d: Unrecognized message type: %v", id, rawMsgType)
		}
	} else {
		logger.Errorf("Node %d: MsgType field not found", id)
	}
}

// startClient connects to another node
func (node *Node) startClient(id, otherID int, address string) (net.Conn, error) {
	conn, err := net.Dial(TcpNetworkType, address)
	if err != nil {
		logger.Errorf("Node %d: Error connecting to Node %d at %s: %v\n", id, otherID, address, err)
		return nil, err
	}

	logger.Infof("Node %d: Connected to Node %d [LOCAL: %s] [REMOTE: %s]\n", id, otherID, conn.LocalAddr(), conn.RemoteAddr())
	return conn, nil
}
