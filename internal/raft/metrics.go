package raft

import (
	"fmt"
	"net/http"
)

func (s *Server) HandleMetrics(
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

	nodeID := s.Node.ID
	state := s.Node.State.String()
	term := s.Node.CurrentTerm
	commitIndex := s.Node.CommitIndex
	lastApplied := s.Node.LastApplied
	lastLogIndex := s.Node.Log.LastIndex()

	s.Node.mu.Unlock()

	leader := 0
	if state == Leader.String() {
		leader = 1
	}

	w.Header().Set(
		"Content-Type",
		"text/plain; version=0.0.4",
	)

	fmt.Fprintf(
		w,
		"# HELP cloudscale_node_info Information about the CloudScale node.\n"+
			"# TYPE cloudscale_node_info gauge\n"+
			"cloudscale_node_info{node_id=\"%s\"} 1\n",
		nodeID,
	)

	fmt.Fprintf(
		w,
		"# HELP cloudscale_raft_term Current Raft term.\n"+
			"# TYPE cloudscale_raft_term gauge\n"+
			"cloudscale_raft_term{node_id=\"%s\"} %d\n",
		nodeID,
		term,
	)

	fmt.Fprintf(
		w,
		"# HELP cloudscale_raft_leader Whether this node is currently leader.\n"+
			"# TYPE cloudscale_raft_leader gauge\n"+
			"cloudscale_raft_leader{node_id=\"%s\"} %d\n",
		nodeID,
		leader,
	)

	fmt.Fprintf(
		w,
		"# HELP cloudscale_raft_commit_index Current committed log index.\n"+
			"# TYPE cloudscale_raft_commit_index gauge\n"+
			"cloudscale_raft_commit_index{node_id=\"%s\"} %d\n",
		nodeID,
		commitIndex,
	)

	fmt.Fprintf(
		w,
		"# HELP cloudscale_raft_last_applied Last log index applied to the state machine.\n"+
			"# TYPE cloudscale_raft_last_applied gauge\n"+
			"cloudscale_raft_last_applied{node_id=\"%s\"} %d\n",
		nodeID,
		lastApplied,
	)

	fmt.Fprintf(
		w,
		"# HELP cloudscale_raft_last_log_index Last log index stored on this node.\n"+
			"# TYPE cloudscale_raft_last_log_index gauge\n"+
			"cloudscale_raft_last_log_index{node_id=\"%s\"} %d\n",
		nodeID,
		lastLogIndex,
	)
}
