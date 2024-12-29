package node

func (n *Node) GetLeaderID() uint64 {
	return (n.View % n.TotalNodes) + 1
}

func (n *Node) IsLeaderNode() bool {
	return n.ID == n.GetLeaderID()
}
