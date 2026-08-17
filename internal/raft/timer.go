package raft

import (
	"time"
)

func (n *Node) ElectionTimerExpired() bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.State == Leader {
		return false
	}

	return time.Since(n.LastHeartbeat) >= n.ElectionTimeout
}