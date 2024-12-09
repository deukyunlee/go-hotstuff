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
	NodeTable map[uint64]string
)

func StartNewNode(nodeId uint64, runningInDocker bool) *Node {
	if runningInDocker {
		NodeTable = map[uint64]string{
			1: "node1:1111",
			2: "node2:1112",
			3: "node3:1113",
			4: "node4:1114",
		}
	} else {
		NodeTable = map[uint64]string{
			1: "localhost:1111",
			2: "localhost:1112",
			3: "localhost:1113",
			4: "localhost:1114",
		}
	}

	return setNode(nodeId, NodeTable[nodeId])
}

// setNode starts both server and client and set data for a given node
func setNode(nodeId uint64, address string) *Node {

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

func (node *Node) startClientWithRetry(nodeId uint64, otherId uint64, otherAddress string) error {
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
func (node *Node) startServer(nodeId uint64, address string) {
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

		node.LocalConnection = conn

		go node.handleConnection(nodeId, conn)
	}
}

func (node *Node) startServerWithRetry(nodeId uint64, ln net.Listener) (conn net.Conn, err error) {
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
func (node *Node) handleConnection(id uint64, conn net.Conn) {
	logger.Infof("Node %d: Accepted connection from %s\n", id, conn.RemoteAddr().String())

	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer)
	logger.Infof("Node %d: Received raw buffer: %s", id, string(buffer[:n]))

	var rawMsg map[string]interface{}
	err = json.Unmarshal(buffer[:n], &rawMsg)
	if err != nil {
		logger.Errorf("Node %d: Error unmarshalling message: %v", id, err)
		return
	}

	logger.Infof("[NodeId: %d] rawMsg: %v", node.NodeID, rawMsg)
	//logger.Infof("rawMsg msgType: %v", rawMsg["msgType"].(int))
	if rawMsgType, ok := rawMsg["msgType"].(float64); ok {

		consensusMsgType := consensus.MsgType(int(rawMsgType))
		switch consensusMsgType {
		case consensus.Request:
			var reqMsg consensus.RequestMsg
			err := json.Unmarshal(buffer[:n], &reqMsg)
			if err != nil {
				logger.Errorf("Node %d: Error unmarshalling message: %v", id, err)
			}
			node.MsgDelivery <- &reqMsg
		case consensus.Prepare:
			var prepareMsg consensus.PrepareMsg
			err := json.Unmarshal(buffer[:n], &prepareMsg)
			if err != nil {
				logger.Errorf("Node %d: Error unmarshalling message: %v", id, err)
			}
			node.MsgDelivery <- &prepareMsg
		case consensus.PreCommit:
			var preCommitMsg consensus.PreCommitMsg
			err := json.Unmarshal(buffer[:n], &preCommitMsg)
			if err != nil {
				logger.Errorf("Node %d: Error unmarshalling message: %v", id, err)
			}
			node.MsgDelivery <- &preCommitMsg
		case consensus.Commit:
			var commitMsg consensus.CommitMsg
			err := json.Unmarshal(buffer[:n], &commitMsg)
			if err != nil {
				logger.Errorf("Node %d: Error unmarshalling message: %v", id, err)
			}
			node.MsgDelivery <- &commitMsg
		case consensus.Decide:
			var decideMsg consensus.DecideMsg
			err := json.Unmarshal(buffer[:n], &decideMsg)
			if err != nil {
				logger.Errorf("Node %d: Error unmarshalling message: %v", id, err)
			}
			node.MsgDelivery <- &decideMsg
		default:
			logger.Errorf("Node %d: Unrecognized message type: %v", id, rawMsgType)
		}
	} else {
		logger.Errorf("Node %d: MsgType field not found with %s", id, rawMsg)
	}

}

// startClient connects to another node
func (node *Node) startClient(id, otherID uint64, address string) (net.Conn, error) {
	conn, err := net.Dial(TcpNetworkType, address)
	if err != nil {
		logger.Errorf("Node %d: Error connecting to Node %d at %s: %v\n", id, otherID, address, err)
		return nil, err
	}

	logger.Infof("Node %d: Connected to Node %d [LOCAL: %s] [REMOTE: %s]\n", id, otherID, conn.LocalAddr(), conn.RemoteAddr())
	return conn, nil
}

func (node *Node) SendInitialRequestMsg() {
	leader := node.CurrentState.GetLeader()

	if node.NodeID != leader {
		reqMsg := consensus.RequestMsg{
			MsgType:    consensus.Request,
			Timestamp:  time.Now().Unix(),
			ClientID:   node.NodeID,
			Operation:  "INITIAL",
			SequenceID: 1,
		}

		conn, err := net.Dial("tcp", NodeTable[leader])
		if err != nil {
			logger.Errorf("Failed to connect to leader: %v", err)
			return
		}
		defer conn.Close()

		jsonReqMsg, _ := json.Marshal(reqMsg)
		_, err = conn.Write(jsonReqMsg)
		if err != nil {
			logger.Errorf("Failed to send request: %v", err)
		}
	}
}
