package types

type QuorumCertificate struct {
    BlockHash    string
    View         uint64
    Phase        MessageType
    Signatures   map[uint64]string  // nodeID -> signature
}