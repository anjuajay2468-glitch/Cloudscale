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
	votes := 1 // vote for ourselves

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

		// Another node has a newer term.
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

	n.mu.Lock()
	defer n.mu.Unlock()

	// We may have stepped down while the election was happening.
	if n.State != Candidate || n.CurrentTerm != term {
		return false
	}

	// Majority of the total cluster.
	majority := n.ClusterSize/2 + 1

	if votes >= majority {
		n.State = Leader
		n.LastHeartbeat = time.Now()
		return true
	}

	return false
}