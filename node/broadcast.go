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
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		logger.Errorf("Failed to marshall msg: %v\n", jsonMsg)
	}

	_, err = n.Connections[id].Write(jsonMsg)
	if err != nil {
		logger.Errorf("Failed to send message to %s: %v", n.Connections[id].RemoteAddr(), err)
	}
}
