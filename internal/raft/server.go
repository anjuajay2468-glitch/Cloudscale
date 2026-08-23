package raft

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"github.com/anjuajay2468-glitch/cloudscale/internal/auth"
)

type Server struct {
	Node          *Node
	ApplyFunc     ApplyFunc
	InternalToken string
}

func NewServer(node *Node) *Server {
	return &Server{
		Node: node,
	}
}

func (s *Server) SetApplyFunc(apply ApplyFunc) {
	s.ApplyFunc = apply
}

func (s *Server) SetInternalToken(token string) {
	s.InternalToken = token
}

// authenticateInternal verifies node-to-node authentication.
//
// If no token is configured, authentication is skipped. This keeps the
// Raft package usable by unit tests. Production nodes configure the token
// during startup.
func (s *Server) authenticateInternal(w http.ResponseWriter, r *http.Request) bool {
	if s.InternalToken == "" {
		return true
	}

	provided := r.Header.Get(auth.InternalTokenHeader)

	if len(provided) != len(s.InternalToken) ||
		subtle.ConstantTimeCompare(
			[]byte(provided),
			[]byte(s.InternalToken),
		) != 1 {
		http.Error(
			w,
			"forbidden",
			http.StatusForbidden,
		)
		return false
	}

	return true
}

func (s *Server) HandleRequestVote(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	if !s.authenticateInternal(w, r) {
		return
	}

	var args RequestVoteArgs

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(
			w,
			"invalid request",
			http.StatusBadRequest,
		)
		return
	}

	reply := s.Node.RequestVote(args)

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if err := json.NewEncoder(w).Encode(reply); err != nil {
		http.Error(
			w,
			"failed to encode response",
			http.StatusInternalServerError,
		)
	}
}

func (s *Server) HandleAppendEntries(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	if !s.authenticateInternal(w, r) {
		return
	}

	var args AppendEntriesArgs

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(
			w,
			"invalid request",
			http.StatusBadRequest,
		)
		return
	}

	reply := s.Node.HandleAppendEntries(args)

	if reply.Success && s.ApplyFunc != nil {
		if err := s.Node.ApplyCommitted(s.ApplyFunc); err != nil {
			http.Error(
				w,
				"failed to apply committed entries",
				http.StatusInternalServerError,
			)
			return
		}
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if err := json.NewEncoder(w).Encode(reply); err != nil {
		http.Error(
			w,
			"failed to encode response",
			http.StatusInternalServerError,
		)
	}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(
		"/raft/request-vote",
		s.HandleRequestVote,
	)

	mux.HandleFunc(
		"/raft/append-entries",
		s.HandleAppendEntries,
	)

	mux.HandleFunc(
		"/raft/status",
		s.HandleStatus,
	)
	mux.HandleFunc(
		"/metrics",
		s.HandleMetrics,
	)
}

func (s *Server) HandleStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	s.Node.mu.Lock()

	replication := make(map[string]interface{})

	for peer, matchIndex := range s.Node.MatchIndex {
		replication[peer] = map[string]int{
			"match_index": matchIndex,
			"next_index":  s.Node.NextIndex[peer],
		}
	}

	status := map[string]interface{}{
		"id":             s.Node.ID,
		"state":          s.Node.State.String(),
		"term":           s.Node.CurrentTerm,
		"voted_for":      s.Node.VotedFor,
		"last_heartbeat": s.Node.LastHeartbeat,
		"last_log_index": s.Node.Log.LastIndex(),
		"commit_index":   s.Node.CommitIndex,
		"last_applied":   s.Node.LastApplied,
		"replication":    replication,
	}

	s.Node.mu.Unlock()

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(
			w,
			"failed to encode status",
			http.StatusInternalServerError,
		)
	}
}
