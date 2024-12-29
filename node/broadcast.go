package node

import (
	"deukyunlee/hotstuff/message"
	"encoding/json"
	"net"
)

func (n *Node) Broadcast(msg message.Message) {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		logger.Errorf("Failed to marshal message: %v\n", err)
		return
	}

	workerCount := uint64(len(n.Connections) - 1)
	jobs := make(chan net.Conn, workerCount)

	for i := uint64(0); i < workerCount; i++ {
		go func() {
			for conn := range jobs {
				if i == n.ID {
					continue
				}
				_, err := conn.Write(jsonMsg)
				if err != nil {
					logger.Errorf("Failed to send message to %s: %v", conn.RemoteAddr(), err)
				}
			}
		}()
	}

	for _, conn := range n.Connections {
		jobs <- conn
	}

	logger.Infof("closing job for Message: %v\n", msg)
	close(jobs)
}

func (n *Node) Unicast(msg message.Message, id uint64) {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		logger.Errorf("Failed to marshall msg: %v\n", jsonMsg)
	}

	_, err = n.Connections[id].Write(jsonMsg)
	if err != nil {
		logger.Errorf("Failed to send message to %s: %v", n.Connections[id].RemoteAddr(), err)
	}
}
