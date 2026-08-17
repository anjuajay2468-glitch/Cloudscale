package raft

import "time"

type AppendEntriesArgs struct {
	Term     int
	LeaderID string
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (n *Node) HandleAppendEntries(args AppendEntriesArgs) AppendEntriesReply {
	n.mu.Lock()
	defer n.mu.Unlock()

	reply := AppendEntriesReply{
		Term:    n.CurrentTerm,
		Success: false,
	}

	// Reject an old leader.
	if args.Term < n.CurrentTerm {
		return reply
	}

	// A newer term means we must follow this leader.
	if args.Term > n.CurrentTerm {
		n.CurrentTerm = args.Term
		n.VotedFor = ""
	}

	n.State = Follower
	n.LastHeartbeat = time.Now()

	reply.Term = n.CurrentTerm
	reply.Success = true

	return reply
}
func (n *Node) SendHeartbeats(peers []string) {
	n.mu.Lock()

	if n.State != Leader {
		n.mu.Unlock()
		return
	}

	args := AppendEntriesArgs{
		Term:     n.CurrentTerm,
		LeaderID: n.ID,
	}

	n.mu.Unlock()

	for _, peer := range peers {
		reply, err := SendAppendEntries(peer, args)

		if err != nil {
			continue
		}

		n.mu.Lock()

		if reply.Term > n.CurrentTerm {
			n.CurrentTerm = reply.Term
			n.State = Follower
			n.VotedFor = ""
		}

		n.mu.Unlock()
	}
}
func (n *Node) HeartbeatLoop(
	peers []string,
	interval time.Duration,
	stop <-chan struct{},
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			n.SendHeartbeats(peers)

		case <-stop:
			return
		}
	}
}