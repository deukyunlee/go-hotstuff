package network

import (
	"deukyunlee/hotstuff/logging"
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

	for otherID, otherAddress := range NodeTable {
		if otherID != nodeId {
			var conn net.Conn
			var err error

			for i := 0; i < RetryCount; i++ {
				conn, err = startClient(nodeId, otherID, otherAddress)

				if err == nil {
					logger.Infof("Node %d: Connected to Node %d [LOCAL: %s] [REMOTE: %s]\n", nodeId, otherID, conn.LocalAddr(), conn.RemoteAddr())
					node.Connections = append(node.Connections, conn)
					break
				}
				logger.Infof("Failed to connect to node %d (attempt %d/%d): %v\n", otherID, i+1, RetryCount, err)

				time.Sleep(ConnectionDelay)
			}

			if err != nil {
				logger.Errorf("Failed to connect to node %d after %d attempts: %v\n", otherID, RetryCount, err)
				panic(err)
			}
		}
	}

	return node
}

// startServer starts the TCP server for the node
func (node *Node) startServer(nodeId int, address string) {
	ln, err := net.Listen(TcpNetworkType, address)
	if err != nil {
		logger.Errorf("Node %d: Error starting server: %v\n", nodeId, err)
	}

	defer ln.Close()
	if NodeTable[nodeId] == "" {
		panic("Unable to get server info")
	}

	for {
		logger.Infof("Waiting for connection on %s", address)

		var conn net.Conn
		var err error

		for i := 0; i < RetryCount; i++ {
			conn, err = ln.Accept()
			if err == nil {
				logger.Errorf("Node %d: Error accepting connection: %v\n", nodeId, err)
				break
			}

			logger.Infof("Failed to connect to node %d (attempt %d/%d): %d\n", nodeId, i+1, RetryCount, err)

			time.Sleep(ConnectionDelay)
		}

		if err != nil {
			logger.Errorf("Failed to connect to node %d after %d attempts: %d\n", nodeId, RetryCount, err)
			panic(err)
		}

		logger.Infof("Connection accepted from %s", conn.RemoteAddr().String())

		node.setRoute(conn)
		go handleConnection(nodeId, conn)
	}
}

// handleConnection handles incoming connections to the server
func handleConnection(id int, conn net.Conn) {
	logger.Infof("Node %d: Accepted connection from %s\n", id, conn.RemoteAddr().String())

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err == nil {
		logger.Errorf("Node %d: Received message: %s\n", id, string(buffer[:n]))
	}
}

// startClient connects to another node
func startClient(id, otherID int, address string) (net.Conn, error) {
	conn, err := net.Dial(TcpNetworkType, address)
	if err != nil {
		logger.Errorf("Node %d: Error connecting to Node %d at %s: %v\n", id, otherID, address, err)
		return nil, err
	}

	logger.Infof("Node %d: Connected to Node %d [LOCAL: %s] [REMOTE: %s]\n", id, otherID, conn.LocalAddr(), conn.RemoteAddr())
	return conn, nil
}
