package raft

import "time"

// StartElection starts a new Raft election.
//
// peers contains the URLs of the other Raft nodes.
// Example:
//   []string{"http://localhost:8002", "http://localhost:8003"}
//
// The node votes for itself and then asks its peers for votes.
func (n *Node) StartElection(peers []string) bool {
	n.mu.Lock()

	n.State = Candidate
	n.CurrentTerm++
	n.VotedFor = n.ID
	n.LastHeartbeat = time.Now()

	term := n.CurrentTerm
	votes := 1

	n.mu.Unlock()

	for _, peer := range peers {
		reply, err := SendRequestVote(
			peer,
			RequestVoteArgs{
				Term:        term,
				CandidateID: n.ID,
			},
		)

		if err != nil {
			continue
		}

		n.mu.Lock()

		if reply.Term > n.CurrentTerm {
			n.CurrentTerm = reply.Term
			n.State = Follower
			n.VotedFor = ""
			n.mu.Unlock()

			return false
		}

		if reply.VoteGranted {
			votes++
		}

		n.mu.Unlock()
	}

	// Check whether we won the election.
	n.mu.Lock()

	if n.State != Candidate || n.CurrentTerm != term {
		n.mu.Unlock()
		return false
	}

	majority := n.ClusterSize/2 + 1

	if votes < majority {
		n.mu.Unlock()
		return false
	}

	// We are now leader.
	n.State = Leader
	n.LastHeartbeat = time.Now()

	n.mu.Unlock()

	// IMPORTANT:
	// InitializeLeaderReplication() acquires n.mu itself,
	// so it must be called after releasing the lock.
	n.InitializeLeaderReplication(peers)

	return true
}
