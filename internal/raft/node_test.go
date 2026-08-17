package raft

import "testing"

func TestNewNode(t *testing.T) {
	node := NewNode("node1")

	if node.ID != "node1" {
		t.Fatalf("expected node1, got %s", node.ID)
	}

	if node.State != Follower {
		t.Fatalf("expected follower state")
	}

	if node.CurrentTerm != 0 {
		t.Fatalf("expected initial term 0")
	}

	if node.VotedFor != "" {
		t.Fatalf("expected no initial vote")
	}
}
func TestRequestVote(t *testing.T) {
	node := NewNode("node1")

	reply := node.RequestVote(RequestVoteArgs{
		Term:        1,
		CandidateID: "node2",
	})

	if !reply.VoteGranted {
		t.Fatal("expected vote to be granted")
	}

	if node.VotedFor != "node2" {
		t.Fatalf("expected vote for node2, got %s", node.VotedFor)
	}

	// Node 1 should not vote for another candidate
	// in the same term.
	reply = node.RequestVote(RequestVoteArgs{
		Term:        1,
		CandidateID: "node3",
	})

	if reply.VoteGranted {
		t.Fatal("node should not grant a second vote in the same term")
	}
}
func TestElectionWithoutPeersFails(t *testing.T) {
	node := NewNode("node1")

	elected := node.StartElection([]string{})

	if elected {
		t.Fatal("node should not become leader without a majority")
	}

	if node.State != Candidate {
		t.Fatalf("expected candidate state, got %s", node.State)
	}
}
func TestElectionMajority(t *testing.T) {
	node := NewNode("node1")

	// We don't have real peers here, so this should not
	// become leader.
	elected := node.StartElection([]string{
		"http://localhost:9998",
		"http://localhost:9999",
	})

	if elected {
		t.Fatal("election should fail when peers are unavailable")
	}

	if node.State != Candidate {
		t.Fatalf("expected candidate state, got %s", node.State)
	}

	if node.CurrentTerm != 1 {
		t.Fatalf("expected term 1, got %d", node.CurrentTerm)
	}
}