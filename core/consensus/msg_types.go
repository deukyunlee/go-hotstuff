package consensus

type MsgType int

const (
	Request MsgType = iota
	PrepareVote
	Prepare
	PreCommitVote
	PreCommit
	CommitVote
	Commit
	NewView
	Decide
)

// NewViewMsg represents a message used to initiate or handle a view change in the HotStuff consensus protocol
type NewViewMsg struct {
	ViewId            int64       // The new view number being proposed for the consensus
	SenderId          int         // The identifier of the node sending the NewView message
	QuorumCertificate *QuorumCert // Quorum certificate proving the validity of the previous phases
	Justification     *PrepareMsg // Optional justification message from the last valid Prepare phase
	Timestamp         int64       // The time the NewView message was created to prevent replay attacks
}

// RequestMsg represents a client request message
type RequestMsg struct {
	MsgType    MsgType `json:"msgType"`    // The type of the message (e.g., Request)
	Timestamp  int64   `json:"timestamp"`  // The time the request was created
	ClientID   uint64  `json:"clientID"`   // Unique identifier for the client
	Operation  string  `json:"operation"`  // The operation requested by the client
	SequenceID uint64  `json:"sequenceID"` // Sequence ID for ordering requests
}

// PrepareMsg represents a message used in the Prepare phase
type PrepareMsg struct {
	ViewID     uint64      `json:"viewID"`     // The current view ID
	SequenceID uint64      `json:"sequenceID"` // Sequence ID for ordering messages
	Digest     string      `json:"digest"`     // Digest of the request message
	RequestMsg *RequestMsg `json:"requestMsg"` // The original request message
	NodeID     uint64      `json:"nodeID"`     // ID of the node sending this message
	Signature  string      `json:"signature"`  // Digital signature for message authenticity
}

// PreCommitMsg represents a message used in the PreCommit phase
type PreCommitMsg struct {
	ViewID     uint64  `json:"viewID"`     // The current view ID
	SequenceID uint64  `json:"sequenceID"` // Sequence ID for ordering messages
	Digest     string  `json:"digest"`     // Digest of the request or state
	NodeID     uint64  `json:"nodeID"`     // ID of the node sending this message
	MsgType    MsgType `json:"msgType"`    // Type of the consensus message
	Signature  string  `json:"signature"`  // Digital signature for message authenticity
}

// CommitMsg represents a message used in the Commit phase
type CommitMsg struct {
	ViewID     uint64  `json:"viewID"`     // The current view ID
	SequenceID uint64  `json:"sequenceID"` // Sequence ID for ordering messages
	Digest     string  `json:"digest"`     // Digest of the request or state
	NodeID     uint64  `json:"nodeID"`     // ID of the node sending this message
	MsgType    MsgType `json:"msgType"`    // Type of the consensus message
	Signature  string  `json:"signature"`  // Digital signature for message authenticity
}

// DecideMsg represents a message used in the DecideMsg phase
type DecideMsg struct {
	ViewID     uint64  `json:"viewID"`     // The current view ID
	SequenceID uint64  `json:"sequenceID"` // Sequence ID for ordering messages
	Digest     string  `json:"digest"`     // Digest of the request or state
	NodeID     uint64  `json:"nodeID"`     // ID of the node sending this message
	MsgType    MsgType `json:"msgType"`    // Type of the consensus message
	Signature  string  `json:"signature"`  // Digital signature for message authenticity
}

// TimeoutMsg represents a message indicating a timeout
type TimeoutMsg struct {
	ViewID     uint64 `json:"viewID"`     // The view ID where the timeout occurred
	SequenceID uint64 `json:"sequenceID"` // Sequence ID for ordering messages
	NodeID     uint64 `json:"nodeID"`     // ID of the node reporting the timeout
	Reason     string `json:"reason"`     // Reason for the timeout
	Timestamp  uint64 `json:"timestamp"`  // The time the timeout occurred
}

type QuorumCert struct {
	BlockHash []byte
	MsgType   MsgType
	viewNum   uint64
	signature []byte
}

type PrepareVoteMsg struct {
	ViewID     uint64      `json:"viewID"`     // The current view ID
	SequenceID uint64      `json:"sequenceID"` // Sequence ID for ordering messages
	Digest     string      `json:"digest"`     // Digest of the request message
	RequestMsg *RequestMsg `json:"requestMsg"` // The original request message
	NodeID     uint64      `json:"nodeID"`     // ID of the node sending this message
	Signature  string      `json:"signature"`  // Digital signature for message authenticity
}

type PreCommitVoteMsg struct {
	ViewID     uint64  `json:"viewID"`     // The current view ID
	SequenceID uint64  `json:"sequenceID"` // Sequence ID for ordering messages
	Digest     string  `json:"digest"`     // Digest of the request or state
	NodeID     uint64  `json:"nodeID"`     // ID of the node sending this message
	MsgType    MsgType `json:"msgType"`    // Type of the consensus message
	Signature  string  `json:"signature"`  // Digital signature for message authenticity
}

type CommitVoteMsg struct {
	ViewID     uint64  `json:"viewID"`     // The current view ID
	SequenceID uint64  `json:"sequenceID"` // Sequence ID for ordering messages
	Digest     string  `json:"digest"`     // Digest of the request or state
	NodeID     uint64  `json:"nodeID"`     // ID of the node sending this message
	MsgType    MsgType `json:"msgType"`    // Type of the consensus message
	Signature  string  `json:"signature"`  // Digital signature for message authenticity
}

type DecideVoteMsg struct {
	ViewID     uint64  `json:"viewID"`     // The current view ID
	SequenceID uint64  `json:"sequenceID"` // Sequence ID for ordering messages
	Digest     string  `json:"digest"`     // Digest of the request or state
	NodeID     uint64  `json:"nodeID"`     // ID of the node sending this message
	MsgType    MsgType `json:"msgType"`    // Type of the consensus message
	Signature  string  `json:"signature"`  // Digital signature for message authenticity
}
