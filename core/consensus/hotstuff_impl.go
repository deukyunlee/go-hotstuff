package consensus

import (
	"deukyunlee/hotstuff/logging"
	"sync"
)

var (
	logger = logging.GetLogger()
)

type State struct {
	ViewID         uint64
	MsgLogs        *MsgLogs
	LastSequenceID uint64
	CurrentStage   Stage
	mu             *sync.Mutex
}

type MsgLogs struct {
	ReqMsg        []*RequestMsg
	PrepareMsgs   map[string]*PrepareMsg
	ConsensusMsgs map[string]*ConsensusMsg
}

type Stage int

const (
	Idle Stage = iota
	Prepared
	PreCommitted
	Committed
	Decided
)

/*
In DiemBFT, which is based on HotStuff, the leader election process is designed to prevent crash failures. If the elected leader fails to broadcast messages properly due to network issues or hacking, it becomes difficult to reach consensus in that round, potentially causing a halt in block production for one round. To address this, only nodes that actively participated in the previous round's consensus are eligible to be elected as leaders. Additionally, if the elected leader fails to broadcast messages correctly, their reputation score is reduced.

HotStuff requires synchronous communication between nodes during the lock-precursor process, which can introduce latency.
*/

func (state *State) GetLeader() uint64 {
	var id uint64

	logger.Infof("leader: %d", state.ViewID)
	id = state.ViewID % 4
	if id == 0 {
		id = 4
	}

	return id
}

func (state *State) StartConsensus(request *RequestMsg) error {

	sequenceId := uint64(0)

	if sequenceId <= state.LastSequenceID {
		sequenceId = state.LastSequenceID + 1
	}

	request.SequenceID = sequenceId

	state.MsgLogs.ReqMsg = append(state.MsgLogs.ReqMsg, request)

	return nil
}

func (state *State) Prepare(prepareMsg *PrepareMsg) (*ConsensusMsg, error) {

	/*
		TODO
		- as a leader // r = LEADER(curView)
			we assume special NEW_VIEW messages from view 0
			wait for (n - f) NEW-VIEW messages: M <- {m | MATCHINGMSG(m, NEW-VIEW, curView - 1)}
			highQc <- (arg max{m.justify.viewNumber}).justify
			curProposal <- CREATELEAF(highQc.node, client's command
			broadcast MSG(PREPARE, curProposal, highQC)
		- as a replica
			wait for message m : MATCHINGMSG(m, PREPARE, CurView) from LEADER(curView)
			if m.node extends from m.justify.node ^
					SAFENODE(m.node, m.justify) then
				send VOTEMSG(PREPARE,m.node,ㅗ) to LEADER(curView)
	*/

	return nil, nil
}

func (state *State) PreCommit(preCommitMsg *ConsensusMsg) (*ConsensusMsg, error) {
	/*
		TODO
		- as a leader // r = LEADER(curView)
			wait for (n - f) votes: V <- {v | MATCHINGMSG(v, PREPARE, curView)}
			prepareQC <- QC(V)
			broadcast MSG(PRE-COMMIT, ㅗ, prepareQC)
		- as a replica
			wait for message m: MATCHINGQC(m.justify, PREPARE, curView) from LEADER(curView)
			prepareQc <- m.justify
			send VOTEMSG(PRE-COMMIT, m.justify.node, ㅗ) to LEADER(curView)
	*/

	return nil, nil
}

func (state *State) Commit(commitMsg *ConsensusMsg) (*ConsensusMsg, error) {
	/*
		TODO
		- as a leader // r = LEADER(curView)
			wait for (n -f) votes: V <- {v | MATCHINGMSG(v, PRE-COMMIT, curView)}
			precommitQC <- QC(V)
			broadcast MSG(COMMIT, ㅗ, precommitQC)

		- as a replica
			wait for message m : MATCHINGQC(m.justify, PRE-COMMIT, curView) from LEADER(curView)
			lockedQC <- m.justify
			send VOTEMSG(COMMIT, m.justify.node, ㅗ) to LEADER(curView)

	*/

	return nil, nil
}

func (state *State) Decide(decideMsg *ConsensusMsg) (*ConsensusMsg, error) {
	/*
		TODO
		- as a leader // r = LEADER(curView)
			wait for (n - f) votes: V <- {v | MATCHINGMSG(v, COMMIT, curView)}
			commitQC <- QC(V)
			broadcast MSG(DECIED, ㅗ, commitQC)

		- as a replica
			wait for message m from LEADER(curView)
			wait for message m : MATCHINGQC(m.justify, COMMIT, curView) from LEADER(curView)
			execute new commands through m.justify.node, respond to clients
	*/

	return nil, nil
}

// Finally
// NEXTVIEW interrupt: goto this line if NEXTVIEW(curView) is called during "wait for" in any phase
// send MSG(NEW-VIEW, ㅗ, prepareQC) to LEADER(curView + 1)

func (state *State) HasQuorum() bool {
	totalReplicas := 4
	quorumSize := (2 * totalReplicas) / 3

	//state.mu.Lock()
	//defer state.mu.Unlock()

	responseCount := len(state.MsgLogs.ReqMsg)
	logger.Infof("[Required: %d], [Got: %d]", quorumSize, responseCount)
	return responseCount >= quorumSize
}
