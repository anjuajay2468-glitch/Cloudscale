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