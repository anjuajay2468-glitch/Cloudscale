package raft

import "time"

type AppendEntriesArgs struct {
	Term         int
	LeaderID     string
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
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

		if err := n.saveStateLocked(); err != nil {
			return reply
		}
	}

	n.State = Follower
	n.LastHeartbeat = time.Now()

	// Validate the previous log entry.
	if args.PrevLogIndex > 0 {
		if args.PrevLogIndex > len(n.Log.Entries) {
			reply.Term = n.CurrentTerm
			return reply
		}

		prevEntry := n.Log.Entries[args.PrevLogIndex-1]

		if prevEntry.Term != args.PrevLogTerm {
			reply.Term = n.CurrentTerm
			return reply
		}
	}

	// Append new entries.
	for _, entry := range args.Entries {
		if entry.Index <= len(n.Log.Entries) {
			existing := n.Log.Entries[entry.Index-1]

			if existing.Term != entry.Term {
				// Remove conflicting entries.
				n.Log.Entries = n.Log.Entries[:entry.Index-1]
				n.Log.Append(entry)
			}
		} else {
			n.Log.Append(entry)
		}
	}
	if len(args.Entries) > 0 && n.LogPath != "" {
		if err := n.Log.Save(n.LogPath); err != nil {
			reply.Term = n.CurrentTerm
			return reply
		}
	}

	// Update commit index.
	// Update and persist commit index.
	if args.LeaderCommit > n.CommitIndex {
		newCommitIndex := args.LeaderCommit

		if newCommitIndex > len(n.Log.Entries) {
			newCommitIndex = len(n.Log.Entries)
		}

		n.CommitIndex = newCommitIndex

		if err := n.saveStateLocked(); err != nil {
			reply.Term = n.CurrentTerm
			reply.Success = false
			return reply
		}
	}

	reply.Term = n.CurrentTerm
	reply.Success = true

	return reply
}
func (n *Node) SendHeartbeats(peers []string) {
	for _, peer := range peers {
		n.mu.Lock()

		if n.State != Leader {
			n.mu.Unlock()
			return
		}

		// If this follower has never been initialized,
		// start sending from the beginning of the log.
		nextIndex, ok := n.NextIndex[peer]
		if !ok || nextIndex < 1 {
			nextIndex = 1
		}

		// Never go beyond the leader's next log position.
		if nextIndex > n.Log.LastIndex()+1 {
			nextIndex = n.Log.LastIndex() + 1
		}

		prevLogIndex := nextIndex - 1
		prevLogTerm := 0

		if prevLogIndex > 0 {
			prevLogTerm = n.Log.Entries[prevLogIndex-1].Term
		}

		// Copy the entries so we can release the mutex
		// before making the network request.
		var entries []LogEntry

		if nextIndex <= n.Log.LastIndex() {
			entries = append(
				[]LogEntry(nil),
				n.Log.Entries[nextIndex-1:]...,
			)
		}

		args := AppendEntriesArgs{
			Term:         n.CurrentTerm,
			LeaderID:     n.ID,
			PrevLogIndex: prevLogIndex,
			PrevLogTerm:  prevLogTerm,
			Entries:      entries,
			LeaderCommit: n.CommitIndex,
		}

		n.mu.Unlock()

		reply, err := SendAppendEntries(peer, args)

		if err != nil {
			continue
		}

		n.mu.Lock()

		// Another node may have a newer term.
		if reply.Term > n.CurrentTerm {
			n.CurrentTerm = reply.Term
			n.State = Follower
			n.VotedFor = ""
			n.mu.Unlock()
			continue
		}

		if reply.Success {
			// The follower successfully matched everything through
			// the previous log index plus all entries we sent.
			lastReplicated := prevLogIndex + len(entries)

			if lastReplicated > n.MatchIndex[peer] {
				n.MatchIndex[peer] = lastReplicated
			}

			n.NextIndex[peer] = lastReplicated + 1
		} else {
			// Follower rejected the entries.
			// Move backwards and retry on the next heartbeat.
			if n.NextIndex[peer] > 1 {
				n.NextIndex[peer]--
			}
		}

		n.mu.Unlock()
	}

	// Check whether the newly replicated entries
	// have reached a majority.
	n.AdvanceCommitIndex()
}
func (n *Node) ConfirmLeadership(peers []string) bool {
	n.mu.Lock()

	if n.State != Leader {
		n.mu.Unlock()
		return false
	}

	term := n.CurrentTerm

	// The leader counts as one successful acknowledgement.
	acks := 1

	// Capture the leader's current log position.
	lastIndex := n.Log.LastIndex()
	lastTerm := n.Log.LastTerm()

	n.mu.Unlock()

	for _, peer := range peers {
		reply, err := SendAppendEntries(
			peer,
			AppendEntriesArgs{
				Term:         term,
				LeaderID:     n.ID,
				PrevLogIndex: lastIndex,
				PrevLogTerm:  lastTerm,
				Entries:      nil,
				LeaderCommit: n.CommitIndex,
			},
		)

		if err != nil {
			continue
		}

		n.mu.Lock()

		// We discovered a newer term.
		if reply.Term > n.CurrentTerm {
			n.CurrentTerm = reply.Term
			n.State = Follower
			n.VotedFor = ""
			n.mu.Unlock()

			return false
		}

		// Leadership must still be valid in the same term.
		if n.State != Leader || n.CurrentTerm != term {
			n.mu.Unlock()
			return false
		}

		if reply.Success {
			acks++
		}

		n.mu.Unlock()
	}

	return acks > n.ClusterSize/2
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
