package consensus

import (
	"deukyunlee/hotstuff/util"
	"fmt"
	"time"
)

type State struct {
	ViewID         int64
	MsgLogs        *MsgLogs
	LastSequenceID int64
	CurrentStage   Stage
}

type MsgLogs struct {
	ReqMsg        *RequestMsg
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

func (state *State) StartConsensus(request *RequestMsg) (*PrepareMsg, error) {

	sequenceId := time.Now().UnixNano()

	if sequenceId <= state.LastSequenceID {
		sequenceId = state.LastSequenceID + 1
	}

	request.SequenceID = sequenceId

	state.MsgLogs.ReqMsg = request

	digest, err := util.Digest(request)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	prepareMsg := &PrepareMsg{
		ViewID:     state.ViewID,
		SequenceID: sequenceId,
		Digest:     digest,
		RequestMsg: request,
	}

	if state.MsgLogs.PrepareMsgs == nil {
		state.MsgLogs.PrepareMsgs = make(map[string]*PrepareMsg)
	}
	state.MsgLogs.PrepareMsgs[digest] = prepareMsg

	state.CurrentStage = Prepared

	return prepareMsg, nil
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
