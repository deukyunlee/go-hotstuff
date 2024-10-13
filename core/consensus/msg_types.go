package consensus

type RequestMsg struct {
	Timestamp  int64  `json:"timestamp"`
	ClientID   string `json:"clientID"`
	Operation  string `json:"operation"`
	SequenceID int64  `json:"sequenceID"`
}

type PrepareMsg struct {
	ViewID     int64       `json:"viewID"`
	SequenceID int64       `json:"sequenceID"`
	Digest     string      `json:"digest"`
	RequestMsg *RequestMsg `json:"requestMsg"`
	NodeID     int         `json:"nodeID"`
	Signature  string      `json:"signature"` // Digital signature for message authenticity
}

type ConsensusMsg struct {
	ViewID     int64   `json:"viewID"`
	SequenceID int64   `json:"sequenceID"`
	Digest     string  `json:"digest"`
	NodeID     int     `json:"nodeID"`
	MsgType    MsgType `json:"msgType"`
	Signature  string  `json:"signature"`
}

type TimeoutMsg struct {
	ViewID     int64  `json:"viewID"`
	SequenceID int64  `json:"sequenceID"`
	NodeID     int    `json:"nodeID"`
	Reason     string `json:"reason"`
}

type MsgType int

const (
	Prepare MsgType = iota
	PreCommit
	Commit
	Decide
)
