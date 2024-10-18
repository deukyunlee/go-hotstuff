package network

import (
	"bytes"
	"deukyunlee/hotstuff/core/consensus"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
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
	NodeID        int
	View          *View
	CurrentState  *consensus.State
	CommittedMsgs []*consensus.RequestMsg
	MsgBuffer     *MsgBuffer
	MsgEntrance   chan interface{}
	MsgDelivery   chan interface{}
	Alarm         chan bool
	Connections   []net.Conn
}

type View struct {
	ID      int64
	Primary string
}

const ResolvingTimeDuration = time.Second

func NewNode(nodeID int) *Node {
	const startViewId = 10000000000

	node := &Node{
		NodeID: nodeID,
		View: &View{
			ID:      startViewId,
			Primary: "1",
		},

		CurrentState:  nil,
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

func (node *Node) Broadcast(msg interface{}, path string) map[int]error {
	errorMap := make(map[int]error)

	for nodeID, url := range NodeTable {
		if nodeID == node.NodeID {
			continue
		}

		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			errorMap[nodeID] = err
			continue
		}

		send(url+path, jsonMsg)
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
			err := node.routeMsg(msg)
			if err != nil {
				fmt.Println(err)
			}
		case <-node.Alarm:
			err := node.routeMsgWhenAlarmed()
			if err != nil {
				fmt.Println(err)
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
	if node.CurrentState == nil {
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
			panic("unhandled default case")
		}
	}

	return nil
}

func (node *Node) sendReqMsgs() {
	if len(node.MsgBuffer.ReqMsgs) > 0 {
		msgs := make([]*consensus.RequestMsg, len(node.MsgBuffer.ReqMsgs))
		copy(msgs, node.MsgBuffer.ReqMsgs)
		node.MsgDelivery <- msgs
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
			node.handleErrors(node.resolveRequestMsg(m))

		case []*consensus.PrepareMsg:
			node.handleErrors(node.resolvePrepareMsg(m))

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
			fmt.Println(err)
		}
	}
}

func send(url string, msg []byte) {
	//TODO: change to TCP socket

	buff := bytes.NewBuffer(msg)
	http.Post("http://"+url, "application/json", buff)
}

func (node *Node) resolvePrepareMsg(msgs []*consensus.PrepareMsg) []error {
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
	errs := make([]error, 0)

	// Resolve messages
	for _, reqMsg := range msgs {
		err := node.GetReq(reqMsg)
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

func (node *Node) hasQuorumForPrepare() bool {
	totalNodes := len(NodeTable)

	//The quorum size is usually calculated as (2/3) * N + 1, where N is the total number of nodes.
	quorumSize := (2 * totalNodes / 3) + 1

	prepareMsgCount := 0
	for _, msg := range node.MsgBuffer.PrepareMsgs {
		if msg.ViewID == node.CurrentState.ViewID && msg.SequenceID == node.CurrentState.LastSequenceID {
			prepareMsgCount++
		}
	}

	return prepareMsgCount >= quorumSize
}

func (node *Node) GetReq(reqMsg *consensus.RequestMsg) error {

	//err := node.createStateForNewConsensus()
	//if err != nil {
	//	return err
	//}

	prePrepareMsg, err := node.CurrentState.StartConsensus(reqMsg)
	if err != nil {
		return err
	}

	if prePrepareMsg != nil {
		node.Broadcast(prePrepareMsg, "/prepare")
	}

	return nil
}

func (node *Node) GetPrepare(prepareMsg *consensus.PrepareMsg) error {

	commitMsg, err := node.CurrentState.Prepare(prepareMsg)
	if err != nil {
		return err
	}

	if commitMsg != nil {
		commitMsg.NodeID = node.NodeID

		if node.hasQuorumForPrepare() {
			node.Broadcast(commitMsg, "/precommit")
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
