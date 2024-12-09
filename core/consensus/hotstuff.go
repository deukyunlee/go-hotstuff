package consensus

type HotStuff interface {
	NewView(newViewMsg *NewViewMsg) error
	StartConsensus(request *RequestMsg) (*PrepareMsg, error)
	Prepare(prepareMsg *PrepareMsg) (*PreCommitMsg, error)
	PreCommit(preCommitMsg *PreCommitMsg) (*CommitMsg, error)
	Commit(commitMsg *CommitMsg) (*DecideMsg, error)
	Decide(decideMsg *DecideMsg) error
}
