package network

import (
	"deukyunlee/hotstuff/core/consensus"
	"deukyunlee/hotstuff/errorhandler"
	"encoding/json"
	"fmt"
	"net/http"
)

type Server struct {
	url  string
	node *Node
}

func NewServer(nodeID string) *Server {
	node := NewNode(nodeID)
	if node.NodeTable[nodeID] == "" {
		panic("Unable to get server info")
	}

	server := &Server{node.NodeTable[nodeID], node}

	server.setRoute()

	return server
}

func (server *Server) Start() {
	fmt.Printf("Starting server %s...\n", server.url)
	if err := http.ListenAndServe(server.url, nil); err != nil {
		fmt.Println(err)
		return
	}
}

func (server *Server) setRoute() {
	http.HandleFunc("/health", server.healthCheck)
	http.HandleFunc("/req", server.getReq)
	http.HandleFunc("/prepare", server.getPrepare)
	http.HandleFunc("/precommit", server.getPreCommit)
	http.HandleFunc("/commit", server.getCommit)
	http.HandleFunc("/decide", server.getDecide)
}

func (server *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Server is healthy!"))

	if err != nil {
		errorhandler.LogError(errorhandler.HTTP_WRITE_CONTENT_ERROR, err)
		return
	}
}

func (server *Server) getReq(writer http.ResponseWriter, request *http.Request) {
	var msg consensus.RequestMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg
}

func (server *Server) getPrepare(writer http.ResponseWriter, request *http.Request) {
	var msg consensus.PrepareMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg
}

func (server *Server) getPreCommit(writer http.ResponseWriter, request *http.Request) {
	var msg consensus.ConsensusMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg
}

func (server *Server) getCommit(writer http.ResponseWriter, request *http.Request) {
	var msg consensus.ConsensusMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg
}

func (server *Server) getDecide(writer http.ResponseWriter, request *http.Request) {
	// TODO: develop
}
