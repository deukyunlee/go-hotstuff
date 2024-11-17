package network

import (
	"deukyunlee/hotstuff/core/consensus"
	"encoding/json"
	"net"
	"time"
)

type MsgBuffer struct {
	ReqMsgs       []*consensus.RequestMsg
	PrepareMsgs   []*consensus.PrepareMsg
	PreCommitMsgs []*consensus.ConsensusMsg
	CommitMsgs    []*consensus.ConsensusMsg
	DecideMsgs    []*consensus.ConsensusMsg
}

type Node struct {
	NodeID          uint64
	View            *View
	CurrentState    *consensus.State
	CommittedMsgs   []*consensus.RequestMsg
	MsgBuffer       *MsgBuffer
	MsgEntrance     chan interface{}
	MsgDelivery     chan interface{}
	Alarm           chan bool
	Connections     []net.Conn
	LocalConnection net.Conn
}

type View struct {
	ID      uint64
	Primary string
}

const ResolvingTimeDuration = time.Second

func NewNode(nodeID uint64) *Node {
	const startViewId = 1

	node := &Node{
		NodeID: nodeID,
		View: &View{
			ID:      startViewId,
			Primary: "1",
		},

		CurrentState: &consensus.State{
			ViewID:         startViewId,
			MsgLogs:        &consensus.MsgLogs{},
			LastSequenceID: 1,
			CurrentStage:   consensus.Idle,
		},
		CommittedMsgs: make([]*consensus.RequestMsg, 0),
		MsgBuffer: &MsgBuffer{
			ReqMsgs:       make([]*consensus.RequestMsg, 0),
			PrepareMsgs:   make([]*consensus.PrepareMsg, 0),
			PreCommitMsgs: make([]*consensus.ConsensusMsg, 0),
			CommitMsgs:    make([]*consensus.ConsensusMsg, 0),
			DecideMsgs:    make([]*consensus.ConsensusMsg, 0),
		},

		MsgEntrance: make(chan interface{}),
		MsgDelivery: make(chan interface{}),
		Alarm:       make(chan bool),
	}

	go node.dispatchMsg()

	go node.alarmToDispatcher()

	go node.resolveMsg()

	return node
}

func (node *Node) Broadcast(msg interface{}) map[int]error {
	errorMap := make(map[int]error)

	for id, conn := range node.Connections {
		if uint64(id) == node.NodeID {
			continue
		}

		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			errorMap[id] = err
			continue
		}

		go func(conn net.Conn, msg []byte) {
			_, err = conn.Write(msg)
			if err != nil {
				logger.Errorf("Failed to send message to %s: %v", conn.RemoteAddr(), err)
			}
		}(conn, jsonMsg)
	}

	if len(errorMap) == 0 {
		return nil
	} else {
		return errorMap
	}
}

func (node *Node) dispatchMsg() {
	for {
		select {
		case msg := <-node.MsgEntrance:
			logger.Infof("[node %d] msEntrance", node.NodeID)
			err := node.routeMsg(msg)
			if err != nil {
				logger.Errorf("Error happend when routing: %v", err)
			}
		case <-node.Alarm:
			err := node.routeMsgWhenAlarmed()
			if err != nil {
				logger.Errorf("Error happend when routing with alarm: %v", err)
			}
		}
	}
}

func (node *Node) routeMsg(msg interface{}) []error {
	switch m := msg.(type) {
	case *consensus.RequestMsg:
		node.processRequestMsg(m)

	case *consensus.PrepareMsg:
		node.processPrepareMsg(m)

	case *consensus.ConsensusMsg:
		switch m.MsgType {
		case consensus.PreCommit:
			node.processConsensusMsg(m, node.MsgBuffer.PreCommitMsgs, consensus.PreCommitted)
		case consensus.Commit:
			node.processConsensusMsg(m, node.MsgBuffer.CommitMsgs, consensus.Committed)
		case consensus.Decide:
			node.processConsensusMsg(m, node.MsgBuffer.DecideMsgs, consensus.Decided)
		default:
			panic("unhandled default case")
		}
	}

	return nil
}

func (node *Node) processRequestMsg(msg *consensus.RequestMsg) {
	if node.CurrentState == nil {
		msgs := append(node.MsgBuffer.ReqMsgs, msg)
		node.MsgBuffer.ReqMsgs = nil
		node.MsgDelivery <- msgs
	} else {
		node.MsgBuffer.ReqMsgs = append(node.MsgBuffer.ReqMsgs, msg)
	}
}

func (node *Node) processPrepareMsg(msg *consensus.PrepareMsg) {
	if node.CurrentState == nil {
		msgs := append(node.MsgBuffer.PrepareMsgs, msg)
		node.MsgBuffer.PrepareMsgs = nil
		node.MsgDelivery <- msgs
	} else {
		node.MsgBuffer.PrepareMsgs = append(node.MsgBuffer.PrepareMsgs, msg)
	}
}

func (node *Node) processConsensusMsg(msg *consensus.ConsensusMsg, buffer []*consensus.ConsensusMsg, expectedStage consensus.Stage) {
	if node.CurrentState == nil || node.CurrentState.CurrentStage != expectedStage {
		node.MsgBuffer.PreCommitMsgs = append(buffer, msg)
	} else {
		msgs := append(buffer, msg)
		node.MsgBuffer.PreCommitMsgs = nil
		node.MsgDelivery <- msgs
	}
}

func (node *Node) routeMsgWhenAlarmed() []error {
	if node.CurrentState == nil || node.CurrentState.CurrentStage == consensus.Idle {
		node.sendReqMsgs()
		node.sendPrepareMsgs()
	} else {
		switch node.CurrentState.CurrentStage {
		case consensus.PreCommitted:
			node.sendConsensusMsgs(node.MsgBuffer.PreCommitMsgs)
		case consensus.Committed:
			node.sendConsensusMsgs(node.MsgBuffer.CommitMsgs)
		case consensus.Decided:
			node.sendConsensusMsgs(node.MsgBuffer.DecideMsgs)
		default:
			logger.Errorf("[Node: %d]unhandled default case: %v", node.NodeID, node.CurrentState.CurrentStage)
			panic("unhandled default case")
		}
	}

	return nil
}

func (node *Node) sendReqMsgs() {
	if len(node.MsgBuffer.ReqMsgs) > 0 {
		logger.Infof("message: %v", node.MsgBuffer.ReqMsgs)
		msgs := make([]*consensus.RequestMsg, len(node.MsgBuffer.ReqMsgs))
		copy(msgs, node.MsgBuffer.ReqMsgs)
		node.MsgDelivery <- msgs
		node.MsgBuffer.ReqMsgs = nil
	}
}

func (node *Node) sendPrepareMsgs() {
	if len(node.MsgBuffer.PrepareMsgs) > 0 {
		msgs := make([]*consensus.PrepareMsg, len(node.MsgBuffer.PrepareMsgs))
		copy(msgs, node.MsgBuffer.PrepareMsgs)
		node.MsgDelivery <- msgs
	}
}

func (node *Node) sendConsensusMsgs(buffer []*consensus.ConsensusMsg) {
	if len(buffer) > 0 {
		msgs := make([]*consensus.ConsensusMsg, len(buffer))
		copy(msgs, buffer)
		node.MsgDelivery <- msgs
	}
}

func (node *Node) alarmToDispatcher() {
	for {
		time.Sleep(ResolvingTimeDuration)
		node.Alarm <- true
	}
}

func (node *Node) resolveMsg() {
	for {
		msgs := <-node.MsgDelivery

		switch m := msgs.(type) {
		case []*consensus.RequestMsg:
			go node.handleErrors(node.resolveRequestMsg(m))

		case []*consensus.PrepareMsg:
			go node.handleErrors(node.resolvePrepareMsg(m))

		case []*consensus.ConsensusMsg:
			if len(m) == 0 {
				continue
			}

			switch m[0].MsgType {
			case consensus.PreCommit:
				node.handleErrors(node.resolvePreCommitMsg(m))
			case consensus.Commit:
				node.handleErrors(node.resolveCommitMsg(m))
			case consensus.Decide:
				node.handleErrors(node.resolveDecideMsg(m))
			default:
				panic("unhandled default case")
			}
		}
	}
}

func (node *Node) handleErrors(errs []error) {
	if len(errs) > 0 {
		for _, err := range errs {
			logger.Errorf("error handling consensus msg: %v", err.Error())
		}
	}
}

func (node *Node) resolvePrepareMsg(msgs []*consensus.PrepareMsg) []error {
	logger.Info("resolvePrepareMsgs")

	errs := make([]error, 0)

	// Resolve messages
	for _, prePrepareMsg := range msgs {
		err := node.GetPrepare(prePrepareMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

func (node *Node) resolveRequestMsg(msgs []*consensus.RequestMsg) []error {
	logger.Info("resolveRequestMsg")

	errs := make([]error, 0)

	// Resolve messages
	for _, reqMsg := range msgs {
		logger.Infof("Request Message: %v", reqMsg)
		node.createStateForNewConsensus(reqMsg)

		// 메시지를 들어올 때마다 처리하도록
		// 버퍼에 쌓을 필요 X
		// 버퍼에 쌓고 1초 주기로 비우는 것보다 그냥 메세지 들어올 때마다 처리하는 것으로 수정
		prePrepareMsg, err := node.CurrentState.StartConsensus(reqMsg)

		logger.Infof("prePrepareMsg: %s", prePrepareMsg)
		// HasQuorum
		err = node.GetReq(reqMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

func (node *Node) resolvePreCommitMsg(msgs []*consensus.ConsensusMsg) []error {
	errs := make([]error, 0)

	// Resolve messages
	for _, prepareMsg := range msgs {
		err := node.GetPreCommit(prepareMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

func (node *Node) resolveCommitMsg(msgs []*consensus.ConsensusMsg) []error {
	errs := make([]error, 0)

	// Resolve messages
	for _, commitMsg := range msgs {
		err := node.GetCommit(commitMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

func (node *Node) resolveDecideMsg(msgs []*consensus.ConsensusMsg) []error {
	errs := make([]error, 0)

	// Resolve messages
	for _, commitMsg := range msgs {
		err := node.GetDecide(commitMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

func (node *Node) GetReq(reqMsg *consensus.RequestMsg) error {

	if reqMsg != nil {
		// Channel to signal when quorum is reached
		quorumChan := make(chan struct{})

		// Goroutine to wait for quorum
		go func() {
			for {
				if node.CurrentState.HasQuorum() {
					close(quorumChan)
					break
				}
				logger.Infof("[Msg: %d]Sleeping for %s", len(node.CurrentState.MsgLogs.PrepareMsgs), "10 * time.Millisecond")
				time.Sleep(10 * time.Millisecond) // Polling interval (adjust as needed)
			}
		}()

		// Wait for quorum
		<-quorumChan

		// Broadcast after quorum is reached
		logger.Info("here")
		node.Broadcast(reqMsg)
	}

	//if prePrepareMsg != nil {
	//	node.Broadcast(prePrepareMsg)
	//}
	return nil
}

func (node *Node) GetPrepare(prepareMsg *consensus.PrepareMsg) error {

	commitMsg, err := node.CurrentState.Prepare(prepareMsg)
	if err != nil {
		return err
	}

	if commitMsg != nil {
		if node.CurrentState.HasQuorum() {
			node.Broadcast(commitMsg)
		} else {

		}
	}

	return nil
}

func (node *Node) GetPreCommit(commitMsg *consensus.ConsensusMsg) error {
	return nil
}

func (node *Node) GetCommit(commitMsg *consensus.ConsensusMsg) error {
	return nil
}

func (node *Node) GetDecide(commitMsg *consensus.ConsensusMsg) error {
	return nil
}

func (node *Node) createStateForNewConsensus(reqMsg *consensus.RequestMsg) {
	if node.CurrentState == nil {
		node.CurrentState = &consensus.State{}
		node.CurrentState.MsgLogs = &consensus.MsgLogs{}
	}

	node.CurrentState.ViewID = node.View.ID
	node.CurrentState.LastSequenceID = reqMsg.SequenceID
	node.CurrentState.CurrentStage = consensus.Idle
}
