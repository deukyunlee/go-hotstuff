package network

import (
	"deukyunlee/hotstuff/core/consensus"
	"deukyunlee/hotstuff/util"
	"encoding/json"
	"net"
	"strconv"
	"sync"
	"time"
)

type Node struct {
	NodeID          uint64
	View            *View
	CurrentState    *consensus.State
	CommittedMsgs   []*consensus.RequestMsg
	MsgDelivery     chan interface{}
	Alarm           chan bool
	Connections     []net.Conn
	LocalConnection net.Conn
	processingView  sync.Map
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
		MsgDelivery:   make(chan interface{}),
		Alarm:         make(chan bool),
	}

	//go node.dispatchMsg()

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
		//case msg := <-node.MsgEntrance:
		//	logger.Infof("[node %d] msEntrance", node.NodeID)
		//	err := node.routeMsg(msg)
		//	if err != nil {
		//		logger.Errorf("Error happend when routing: %v", err)
		//	}
		//case <-node.Alarm:
		//	err := node.routeMsgWhenAlarmed()
		//	if err != nil {
		//		logger.Errorf("Error happend when routing with alarm: %v", err)
		//	}
		}
	}
}

func (node *Node) routeMsg(msg interface{}) error {
	switch m := msg.(type) {
	case *consensus.RequestMsg:
		node.processRequestMsg(m)
	case *consensus.PrepareMsg:
		node.processPrepareMsg(m)
	case *consensus.PreCommitMsg:
		node.processPreCommitMsg(m)
	case *consensus.CommitMsg:
		node.processCommitMsg(m)
	case *consensus.DecideMsg:
		node.processDecideMsg(m)
	}

	return nil
}

func (node *Node) processRequestMsg(msg *consensus.RequestMsg) {
	node.MsgDelivery <- msg
}

func (node *Node) processPrepareMsg(msg *consensus.PrepareMsg) {
	node.MsgDelivery <- msg
}

func (node *Node) processPreCommitMsg(msg *consensus.PreCommitMsg) {
	node.MsgDelivery <- msg
}

func (node *Node) processCommitMsg(msg *consensus.CommitMsg) {
	node.MsgDelivery <- msg
}

func (node *Node) processDecideMsg(msg *consensus.DecideMsg) {
	node.MsgDelivery <- msg
}

func (node *Node) routeMsgWhenAlarmed() []error {
	//if node.CurrentState == nil || node.CurrentState.CurrentStage == consensus.Idle {
	//	node.sendReqMsgs()
	//	node.sendPrepareMsgs()
	//} else {
	//	switch node.CurrentState.CurrentStage {
	//	case consensus.PreCommitted:
	//		node.sendConsensusMsgs(node.MsgBuffer.PreCommitMsgs)
	//	case consensus.Committed:
	//		node.sendConsensusMsgs(node.MsgBuffer.CommitMsgs)
	//	case consensus.Decided:
	//		node.sendConsensusMsgs(node.MsgBuffer.DecideMsgs)
	//	default:
	//		logger.Errorf("[Node: %d]unhandled default case: %v", node.NodeID, node.CurrentState.CurrentStage)
	//		panic("unhandled default case")
	//	}
	//}

	return nil
}

//func (node *Node) sendReqMsgs() {
//	if len(node.MsgBuffer.ReqMsgs) > 0 {
//		logger.Infof("message: %v", node.MsgBuffer.ReqMsgs)
//		msgs := make([]*consensus.RequestMsg, len(node.MsgBuffer.ReqMsgs))
//		copy(msgs, node.MsgBuffer.ReqMsgs)
//		node.MsgDelivery <- msgs
//		node.MsgBuffer.ReqMsgs = nil
//	}
//}
//
//func (node *Node) sendPrepareMsgs() {
//	if len(node.MsgBuffer.PrepareMsgs) > 0 {
//		msgs := make([]*consensus.PrepareMsg, len(node.MsgBuffer.PrepareMsgs))
//		copy(msgs, node.MsgBuffer.PrepareMsgs)
//		node.MsgDelivery <- msgs
//	}
//}
//
//func (node *Node) sendConsensusMsgs(buffer []*consensus.ConsensusMsg) {
//	if len(buffer) > 0 {
//		msgs := make([]*consensus.ConsensusMsg, len(buffer))
//		copy(msgs, buffer)
//		node.MsgDelivery <- msgs
//	}
//}

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
		case *consensus.RequestMsg:
			go func() {
				err := node.resolveRequestMsg(m)
				if err != nil {
					logger.Errorf("err while resolving request msg: %v", err)
					panic(err)
				}
			}()

		case *consensus.PrepareMsg:
			go func() {
				err := node.resolvePrepareMsg(m)
				if err != nil {
					logger.Errorf("err while resolving prepare msg: %v", err)
					panic(err)
				}
			}()

		case *consensus.PreCommitMsg:
			go func() {
				err := node.resolvePreCommitMsg(m)
				if err != nil {
					logger.Errorf("err while resolving precommit msg: %v", err)
					panic(err)
				}
			}()

		case *consensus.CommitMsg:
			go func() {
				err := node.resolveCommitMsg(m)
				if err != nil {
					logger.Errorf("err while resolving commit msg: %v", err)
					panic(err)
				}
			}()
		case *consensus.DecideMsg:
			go func() {
				err := node.resolveDecideMsg(m)
				if err != nil {
					logger.Errorf("err while resolving decide msg: %v", err)
					panic(err)
				}
			}()

		default:
			panic("unhandled message type")
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

func (node *Node) resolvePrepareMsg(prepareMsg *consensus.PrepareMsg) error {
	logger.Info("resolvePrepareMsg")

	err := node.GetPrepare(prepareMsg)
	if err != nil {
		return err
	}

	return nil
}

func (node *Node) resolveRequestMsg(reqMsg *consensus.RequestMsg) error {
	logger.Info("resolveRequestMsg")

	// Resolve messages
	if _, loaded := node.processingView.LoadOrStore(reqMsg.SequenceID, struct{}{}); loaded {
		logger.Infof("[Client: %d][SequenceId: %d] Already being processed", reqMsg.ClientID, reqMsg.SequenceID)
	} else {
		node.createStateForNewConsensus(reqMsg)
	}
	logger.Infof("Request Message: %v", reqMsg)

	node.CurrentState.StartConsensus(reqMsg)

	err := node.GetReq(reqMsg)
	if err != nil {
		return err
	}

	return nil
}

func (node *Node) resolvePreCommitMsg(prepareMsg *consensus.PreCommitMsg) error {

	err := node.GetPreCommit(prepareMsg)
	if err != nil {
		return err
	}

	return nil
}

func (node *Node) resolveCommitMsg(commitMsg *consensus.CommitMsg) []error {

	err := node.GetCommit(commitMsg)
	if err != nil {
		panic(err)
	}

	return nil
}

func (node *Node) resolveDecideMsg(decideMsg *consensus.DecideMsg) []error {

	err := node.GetDecide(decideMsg)
	if err != nil {
		panic(err)
	}

	return nil
}

func (node *Node) GetReq(reqMsg *consensus.RequestMsg) error {
	//if _, loaded := node.processingView.LoadOrStore(reqMsg.SequenceID, struct{}{}); loaded {
	//	logger.Infof("[Client: %d][SequenceId: %d] Already being processed", reqMsg.ClientID, reqMsg.SequenceID)
	//	return nil
	//}

	quorumChan := make(chan struct{})

	go func() {
		for {
			if node.CurrentState.HasQuorum() {
				close(quorumChan)
				break
			}
			logger.Infof("[Msg: %d]Sleeping for %s", len(node.CurrentState.MsgLogs.PrepareMsgs), "1000 * time.Millisecond")
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	<-quorumChan

	firstHash := util.Hash([]byte(node.CurrentState.MsgLogs.ReqMsg[0].Operation))

	for _, msg := range node.CurrentState.MsgLogs.ReqMsg {
		currentHash := util.Hash([]byte(msg.Operation))
		if firstHash != currentHash {
			panic("wrong hash")
		} else {
			logger.Info("Hash of Operation is all the same")
		}
	}

	digest, err := util.Digest(reqMsg)
	if err != nil {
		logger.Errorf("digest error: %s", err)
		return err
	}

	leaderSig := util.Hash([]byte(strconv.FormatUint(node.NodeID, 10)))

	prepareMsg := &consensus.PrepareMsg{
		ViewID:     node.CurrentState.ViewID,
		SequenceID: reqMsg.SequenceID,
		Digest:     digest,
		RequestMsg: reqMsg,
		NodeID:     node.NodeID,
		Signature:  leaderSig,
	}

	// TODO: replica에서 Leader의 signature가 맞는지 확인하기
	if node.CurrentState.MsgLogs.PrepareMsgs == nil {
		node.CurrentState.MsgLogs.PrepareMsgs = []*consensus.PrepareMsg{}
	}

	node.CurrentState.MsgLogs.PrepareMsgs = append(node.CurrentState.MsgLogs.PrepareMsgs, prepareMsg)
	node.CurrentState.CurrentStage = consensus.Prepared

	node.Broadcast(prepareMsg)

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

func (node *Node) GetPreCommit(commitMsg *consensus.PreCommitMsg) error {
	return nil
}

func (node *Node) GetCommit(commitMsg *consensus.CommitMsg) error {
	return nil
}

func (node *Node) GetDecide(commitMsg *consensus.DecideMsg) error {
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
