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
	NodeID     string      `json:"nodeID"`
}

type ConsensusMsg struct {
	ViewID     int64   `json:"viewID"`
	SequenceID int64   `json:"sequenceID"`
	Digest     string  `json:"digest"`
	NodeID     string  `json:"nodeID"`
	MsgType    MsgType `json:"msgType"`
}

type TimeoutMsg struct {
	ViewID     int64  `json:"viewID"`
	SequenceID int64  `json:"sequenceID"`
	NodeID     string `json:"nodeID"`
}

type MsgType int

const (
	Prepare MsgType = iota
	PreCommit
	Commit
	Decide
)
