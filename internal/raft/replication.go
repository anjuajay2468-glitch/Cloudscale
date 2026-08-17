package raft

func (n *Node) InitializeLeaderReplication(peers []string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	next := n.Log.LastIndex() + 1

	for _, peer := range peers {
		n.NextIndex[peer] = next
		n.MatchIndex[peer] = 0
	}
}
func (n *Node) HasMajority(index int) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	count := 1 // leader itself

	for _, matchIndex := range n.MatchIndex {
		if matchIndex >= index {
			count++
		}
	}

	return count > n.ClusterSize/2
}
func (n *Node) AdvanceCommitIndex() {
	n.mu.Lock()
	defer n.mu.Unlock()

	for index := n.CommitIndex + 1; index <= n.Log.LastIndex(); index++ {
		if n.HasMajorityUnsafe(index) {
			n.CommitIndex = index
		}
	}
}
func (n *Node) HasMajorityUnsafe(index int) bool {
	count := 1

	for _, matchIndex := range n.MatchIndex {
		if matchIndex >= index {
			count++
		}
	}

	return count > n.ClusterSize/2
}