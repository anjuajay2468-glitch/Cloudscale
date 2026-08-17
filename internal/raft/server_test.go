package raft

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestVoteHTTP(t *testing.T) {
	node := NewNode("node1")
	server := NewServer(node)

	handler := http.HandlerFunc(server.HandleRequestVote)

	requestBody := `{
		"Term": 1,
		"CandidateID": "node2"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/raft/request-vote",
		strings.NewReader(requestBody),
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			rec.Code,
		)
	}

	if !strings.Contains(rec.Body.String(), `"VoteGranted":true`) {
		t.Fatalf(
			"expected granted vote, got %s",
			rec.Body.String(),
		)
	}
}
func TestThreeNodeElection(t *testing.T) {
	node1 := NewNode("node1")
	node2 := NewNode("node2")
	node3 := NewNode("node3")

	node1.ClusterSize = 3
	node2.ClusterSize = 3
	node3.ClusterSize = 3

	server2 := NewServer(node2)
	server3 := NewServer(node3)

	httpServer2 := httptest.NewServer(
		http.HandlerFunc(server2.HandleRequestVote),
	)
	defer httpServer2.Close()

	httpServer3 := httptest.NewServer(
		http.HandlerFunc(server3.HandleRequestVote),
	)
	defer httpServer3.Close()

	peers := []string{
		httpServer2.URL,
		httpServer3.URL,
	}

	elected := node1.StartElection(peers)

	if !elected {
		t.Fatal("node1 should have won the election")
	}

	if node1.State != Leader {
		t.Fatalf(
			"expected node1 to become leader, got %s",
			node1.State,
		)
	}

	if node1.CurrentTerm != 1 {
		t.Fatalf(
			"expected term 1, got %d",
			node1.CurrentTerm,
		)
	}

	if node2.VotedFor != "node1" {
		t.Fatalf(
			"expected node2 to vote for node1, got %s",
			node2.VotedFor,
		)
	}

	if node3.VotedFor != "node1" {
		t.Fatalf(
			"expected node3 to vote for node1, got %s",
			node3.VotedFor,
		)
	}
}