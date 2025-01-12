package types

type MessageType int

const (
    Prepare MessageType = iota
    PreCommit
    Commit
    Decide
) 