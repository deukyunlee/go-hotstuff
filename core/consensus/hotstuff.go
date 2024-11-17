package consensus

type HotStuff interface {
	NewView(newViewMsg *NewViewMsg) error
	StartConsensus(request *RequestMsg) (*PrepareMsg, error)
	Prepare(prepareMsg *PrepareMsg) (*ConsensusMsg, error)
	PreCommit(preCommitMsg *ConsensusMsg) (*ConsensusMsg, error)
	Commit(commitMsg *ConsensusMsg) (*ConsensusMsg, error)
	Decide(decideMsg *ConsensusMsg) (*ConsensusMsg, error)
}
