package network

import (
	"bufio"
	"deukyunlee/hotstuff/logging"
	"fmt"
	"net"
	"strings"
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

type Network struct {
	Nodes []*Node
}

var NodeTable = map[int]string{
	1: "localhost:1111",
	2: "localhost:1112",
	3: "localhost:1113",
	4: "localhost:1114",
}

func StartNewServer(nodeId int) {
	startNode(nodeId, NodeTable[nodeId])
}

// startNode starts both server and client for a given node
func startNode(nodeId int, address string) *Node {

	node := NewNode(nodeId)

	go startServer(nodeId, address)

	// Delay to ensure server is up before clients attempt connection
	time.Sleep(1 * time.Second)

	for otherID, otherAddress := range NodeTable {
		if otherID != nodeId {
			conn := startClient(nodeId, otherID, otherAddress)

			node.Connections = append(node.Connections, conn)
		}
	}

	fmt.Println(node.Connections)
	return node
}

// startServer starts the TCP server for the node
func startServer(nodeId int, address string) {
	ln, err := net.Listen(TcpNetworkType, address)
	if err != nil {
		logger.Errorf("Node %d: Error starting server: %v\n", nodeId, err)
	}

	defer ln.Close()
	if NodeTable[nodeId] == "" {
		panic("Unable to get server info")
	}

	//server := &Server{NodeTable[nodeId]}

	for {
		logger.Infof("Waiting for connection on %s", address)

		conn, err := ln.Accept()
		if err != nil {
			logger.Errorf("Node %d: Error accepting connection: %v\n", nodeId, err)
			continue
		}
		logger.Infof("Connection accepted from %s", conn.RemoteAddr().String())
		//server.setRoute(conn)

		go handleConnection(nodeId, conn)
	}
	//return node
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
func startClient(id, otherID int, address string) net.Conn {
	conn, err := net.Dial(TcpNetworkType, address)
	if err != nil {
		logger.Errorf("Node %d: Error connecting to Node %d at %s: %v\n", id, otherID, address, err)
		return nil
	}

	logger.Infof("Node %d: Connected to Node %d [LOCAL: %s] [REMOTE: %s]\n", id, otherID, conn.LocalAddr(), conn.RemoteAddr())
	return conn
}

func (server *Server) setRoute(conn net.Conn) {
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		msg := strings.TrimSpace(scanner.Text())
		logger.Info(msg)

		server.node.MsgEntrance <- &msg
	}
}
