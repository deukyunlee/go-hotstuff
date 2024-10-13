package network

import (
	"bufio"
	"deukyunlee/hotstuff/logging"
	"net"
	"strings"
	"sync"
	"time"
)

const TcpNetworkType string = "tcp"

var (
	logger = logging.GetLogger()
)

type Server struct {
	url  string
	node *Node
}

var NodeTable = map[int]string{
	1: "localhost:1111",
	2: "localhost:1112",
	3: "localhost:1113",
	4: "localhost:1114",
}

func StartNewServer() {

	var wg sync.WaitGroup

	for id, address := range NodeTable {
		wg.Add(1)
		go startNode(id, address, &wg)
	}
}

// startServer starts the TCP server for the node
func startServer(nodeId int, address string) {
	ln, err := net.Listen(TcpNetworkType, address)
	if err != nil {
		logger.Errorf("Node %d: Error starting server: %v\n", nodeId, err)
		return
	}
	defer func(ln net.Listener) {
		err := ln.Close()
		if err != nil {
			logger.Errorf("Node %d: Error closing listener: %v\n", nodeId, err)
		}
	}(ln)

	node := NewNode(nodeId)
	if node.NodeTable[nodeId] == "" {
		panic("Unable to get server info")
	}

	server := &Server{node.NodeTable[nodeId], node}

	logger.Infof("Node %d: Listening on %s\n", nodeId, address)
	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.Errorf("Node %d: Error accepting connection: %v\n", nodeId, err)
			continue
		}
		server.setRoute(conn)
		go handleConnection(nodeId, conn)
	}
}

// handleConnection handles incoming connections to the server
func handleConnection(id int, conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Errorf("Node %d: Error closing connection: %v\n", id, err)
		}
	}(conn)
	logger.Infof("Node %d: Accepted connection from %s\n", id, conn.RemoteAddr().String())

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err == nil {
		logger.Errorf("Node %d: Received message: %s\n", id, string(buffer[:n]))
	}
}

// startNode starts both server and client for a given node
func startNode(nodeId int, address string, wg *sync.WaitGroup) {
	defer wg.Done()

	go startServer(nodeId, address)

	// Delay to ensure server is up before clients attempt connection
	time.Sleep(1 * time.Second)

	for otherID, otherAddress := range NodeTable {
		if otherID != nodeId {
			startClient(nodeId, otherID, otherAddress)
		}
	}
}

// startClient connects to another node
func startClient(id, otherID int, address string) {
	conn, err := net.Dial(TcpNetworkType, address)
	if err != nil {
		logger.Errorf("Node %d: Error connecting to Node %d at %s: %v\n", id, otherID, address, err)
		return
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Errorf("Node %d: Error closing connection: %v\n", id, err)
		}
	}(conn)

	logger.Infof("Node %d: Connected to Node %d at %s\n", id, otherID, address)

	//_, err = conn.Write([]byte("req"))
	//if err != nil {
	//	logger.Errorf("Node %d: Error sending message to Node %d: %v\n", id, otherID, err)
	//}
}

func (server *Server) setRoute(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Errorf("Failed to close connection on %s, %s", conn.RemoteAddr().String(), err.Error())
		}
	}(conn)

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		msg := strings.TrimSpace(scanner.Text())
		logger.Info(msg)

		server.node.MsgEntrance <- &msg
	}
}
