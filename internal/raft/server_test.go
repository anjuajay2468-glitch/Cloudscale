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