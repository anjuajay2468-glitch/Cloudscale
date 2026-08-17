package raft

type ApplyFunc func(entry LogEntry) error

func (n *Node) ApplyCommitted(apply ApplyFunc) error {
	n.mu.Lock()

	for n.LastApplied < n.CommitIndex {
		nextIndex := n.LastApplied + 1

		entry := n.Log.Entries[nextIndex-1]

		n.mu.Unlock()

		if err := apply(entry); err != nil {
			return err
		}

		n.mu.Lock()
		n.LastApplied = nextIndex
	}

	n.mu.Unlock()

	return nil
}