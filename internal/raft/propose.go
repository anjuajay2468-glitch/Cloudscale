package raft

import "fmt"

func (n *Node) Propose(
	command string,
	key string,
	data []byte,
) (LogEntry, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.State != Leader {
		return LogEntry{}, fmt.Errorf("node %s is not leader", n.ID)
	}

	entry := LogEntry{
		Index:   n.Log.LastIndex() + 1,
		Term:    n.CurrentTerm,
		Command: command,
		Key:     key,
		Data:    data,
	}

	n.Log.Append(entry)

	return entry, nil
}
