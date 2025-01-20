package node

import (
	"deukyunlee/hotstuff/message"
	"encoding/binary"
	"encoding/json"
)

func (n *Node) Broadcast(msg message.Message) {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		logger.Errorf("Failed to marshal message: %v\n", err)
		return
	}

	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(jsonMsg)))

	fullMsg := append(lengthBuf, jsonMsg...)

	for targetId, conn := range n.Connections {
		logger.Infof("Broadcasting message to Node %d\n", targetId)
		if targetId == n.ID {
			continue
		}

		_, err := conn.Write(fullMsg)
		if err != nil {
			logger.Errorf("Failed to send message to Node %d: %v", targetId, err)
		}
	}
}

func (n *Node) Unicast(msg message.Message, id uint64) {
	conn, exists := n.Connections[id]
	if !exists {
		logger.Errorf("No connection found for node %d", id)
		return
	}
	
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		logger.Errorf("Failed to marshal message: %v", err)
		return
	}
	
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, uint32(len(msgBytes)))
	
	fullMsg := append(lengthBytes, msgBytes...)
	
	_, err = conn.Write(fullMsg)
	if err != nil {
		logger.Errorf("Failed to send message to node %d: %v", id, err)
		return
	}
	
	logger.Infof("Successfully sent message to node %d", id)
}
