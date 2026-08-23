package raft

import (
	"encoding/json"
	"fmt"
	"os"
)

type persistedState struct {
	CurrentTerm int    `json:"current_term"`
	VotedFor    string `json:"voted_for"`
	CommitIndex int    `json:"commit_index"`
}

func (n *Node) saveStateLocked() error {
	if n.StatePath == "" {
		return nil
	}

	state := persistedState{
		CurrentTerm: n.CurrentTerm,
		VotedFor:    n.VotedFor,
		CommitIndex: n.CommitIndex,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode raft state: %w", err)
	}

	tempPath := n.StatePath + ".tmp"

	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write raft state: %w", err)
	}

	if err := os.Rename(tempPath, n.StatePath); err != nil {
		return fmt.Errorf("failed to replace raft state: %w", err)
	}

	return nil
}

func (n *Node) LoadState(path string) error {
	data, err := os.ReadFile(path)

	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to read raft state: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	var state persistedState

	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to decode raft state: %w", err)
	}

	n.CurrentTerm = state.CurrentTerm
	n.VotedFor = state.VotedFor

	// A persisted commit index can never be beyond the
	// log that was actually recovered from disk.
	n.CommitIndex = state.CommitIndex

	if n.CommitIndex > n.Log.LastIndex() {
		n.CommitIndex = n.Log.LastIndex()
	}

	if n.CommitIndex < 0 {
		n.CommitIndex = 0
	}

	// LastApplied must also never exceed the recovered
	// committed portion of the log.
	n.LastApplied = n.CommitIndex

	return nil
}

func (n *Node) PersistState() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.saveStateLocked()
}
