package raft

import (
	"math/rand"
	"sync"
	"time"
)

// StartElection starts a new Raft election.
//
// Vote requests are sent concurrently so that a failed/unreachable
// peer cannot block the election from reaching healthy peers.
func (n *Node) StartElection(peers []string) bool {
	n.mu.Lock()

	n.State = Candidate
	n.CurrentTerm++
	n.VotedFor = n.ID

	n.ElectionTimeout = time.Duration(
		250+rand.Intn(200),
	) * time.Millisecond

	n.LastHeartbeat = time.Now()

	term := n.CurrentTerm
	votes := 1
	majority := n.ClusterSize/2 + 1

	if err := n.saveStateLocked(); err != nil {
		n.mu.Unlock()
		return false
	}

	// Capture our log position for the RequestVote RPC.
	lastLogIndex := n.Log.LastIndex()
	lastLogTerm := n.Log.LastTerm()

	n.mu.Unlock()

	// Single-node cluster.
	if votes >= majority {
		n.mu.Lock()

		if n.State == Candidate && n.CurrentTerm == term {
			n.State = Leader
			n.LastHeartbeat = time.Now()
			_ = n.saveStateLocked()

			n.mu.Unlock()

			n.InitializeLeaderReplication(peers)

			return true
		}

		n.mu.Unlock()
		return false
	}

	// Vote requests must happen concurrently.
	//
	// If one peer is dead, a slow request to that peer must
	// not prevent us from receiving a vote from a healthy peer.
	type voteResult struct {
		reply RequestVoteReply
		err   error
	}

	results := make(chan voteResult, len(peers))

	var wg sync.WaitGroup

	for _, peer := range peers {
		wg.Add(1)

		go func(peer string) {
			defer wg.Done()

			reply, err := SendRequestVote(
				peer,
				RequestVoteArgs{
					Term:         term,
					CandidateID:  n.ID,
					LastLogIndex: lastLogIndex,
					LastLogTerm:  lastLogTerm,
				},
			)

			results <- voteResult{
				reply: reply,
				err:   err,
			}
		}(peer)
	}

	// Close the results channel after every request completes.
	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result.err != nil {
			continue
		}

		n.mu.Lock()

		// A newer term supersedes this election.
		if result.reply.Term > n.CurrentTerm {
			n.CurrentTerm = result.reply.Term
			n.State = Follower
			n.VotedFor = ""
			n.LastHeartbeat = time.Now()

			_ = n.saveStateLocked()

			n.mu.Unlock()
			continue
		}

		// Ignore stale responses.
		if n.State != Candidate || n.CurrentTerm != term {
			n.mu.Unlock()
			continue
		}

		if result.reply.VoteGranted {
			votes++
		}

		// We have a majority.
		if votes >= majority {
			n.State = Leader
			n.LastHeartbeat = time.Now()

			if err := n.saveStateLocked(); err != nil {
				n.mu.Unlock()
				return false
			}

			n.mu.Unlock()

			n.InitializeLeaderReplication(peers)

			return true
		}

		n.mu.Unlock()
	}

	// Election failed.
	n.mu.Lock()

	if n.State == Candidate && n.CurrentTerm == term {
		n.LastHeartbeat = time.Now()
	}

	n.mu.Unlock()

	return false
}
