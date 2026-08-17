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