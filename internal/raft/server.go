package raft

import (
	"encoding/json"
	"net/http"
)

type Server struct {
	Node *Node
}

func NewServer(node *Node) *Server {
	return &Server{
		Node: node,
	}
}

func (s *Server) HandleRequestVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var args RequestVoteArgs

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	reply := s.Node.RequestVote(args)

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(reply); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
func (s *Server) HandleAppendEntries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var args AppendEntriesArgs

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	reply := s.Node.HandleAppendEntries(args)

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(reply); err != nil {
		http.Error(
			w,
			"failed to encode response",
			http.StatusInternalServerError,
		)
	}
}
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/raft/request-vote", s.HandleRequestVote)
	mux.HandleFunc("/raft/append-entries", s.HandleAppendEntries)
	mux.HandleFunc("/raft/status", s.HandleStatus)
}
func (s *Server) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.Node.mu.Lock()

	status := map[string]interface{}{
		"id":           s.Node.ID,
		"state":        s.Node.State.String(),
		"term":         s.Node.CurrentTerm,
		"voted_for":    s.Node.VotedFor,
		"last_heartbeat": s.Node.LastHeartbeat,
	}

	s.Node.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(
			w,
			"failed to encode status",
			http.StatusInternalServerError,
		)
	}
}