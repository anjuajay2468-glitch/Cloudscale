package raft

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type State int

const (
	Follower State = iota
	Candidate
	Leader
)

type Node struct {
	mu sync.Mutex

	ID string

	State State

	CurrentTerm int

	VotedFor string

	LastHeartbeat time.Time

	ClusterSize int

	ElectionTimeout time.Duration

	Log *Log

	LogPath   string
	StatePath string

	CommitIndex int

	LastApplied int

	NextIndex  map[string]int
	MatchIndex map[string]int
}

func NewNode(id string, clusterSize int) *Node {
	return &Node{
		ID:              id,
		State:           Follower,
		CurrentTerm:     0,
		VotedFor:        "",
		LastHeartbeat:   time.Now(),
		ClusterSize:     clusterSize,
		ElectionTimeout: time.Duration(250+rand.Intn(200)) * time.Millisecond,

		Log:         NewLog(),
		CommitIndex: 0,
		LastApplied: 0,

		NextIndex:  make(map[string]int),
		MatchIndex: make(map[string]int),
	}
}

func NewNodeWithLogPath(
	id string,
	clusterSize int,
	logPath string,
	statePath string,
) (*Node, error) {
	node := NewNode(id, clusterSize)

	node.LogPath = logPath
	node.StatePath = statePath

	if err := node.Log.Load(logPath); err != nil {
		return nil, err
	}

	if err := node.LoadState(statePath); err != nil {
		return nil, err
	}

	// Persist the recovered, internally consistent state.
	node.mu.Lock()
	if err := node.saveStateLocked(); err != nil {
		node.mu.Unlock()
		return nil, err
	}
	node.mu.Unlock()

	return node, nil
}

func (n *Node) PersistLog() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.LogPath == "" {
		return nil
	}

	return n.Log.Save(n.LogPath)
}

func (s State) String() string {
	switch s {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Unknown"
	}
}

func (n *Node) RequestVote(args RequestVoteArgs) RequestVoteReply {
	n.mu.Lock()
	defer n.mu.Unlock()

	reply := RequestVoteReply{
		Term:        n.CurrentTerm,
		VoteGranted: false,
	}

	// Reject candidates from older terms.
	if args.Term < n.CurrentTerm {
		return reply
	}

	// A newer term always supersedes our current state.
	if args.Term > n.CurrentTerm {
		n.CurrentTerm = args.Term
		n.State = Follower
		n.VotedFor = ""
	}

	reply.Term = n.CurrentTerm

	// Only vote for a candidate whose log is at least as
	// up-to-date as our own log.
	//
	// Raft compares the last log term first. If the terms are
	// equal, the candidate with the longer log is more up-to-date.
	candidateUpToDate :=
		args.LastLogTerm > n.Log.LastTerm() ||
			(args.LastLogTerm == n.Log.LastTerm() &&
				args.LastLogIndex >= n.Log.LastIndex())

	if !candidateUpToDate {
		return reply
	}
	// We may vote once per term.
	if n.VotedFor == "" || n.VotedFor == args.CandidateID {
		n.VotedFor = args.CandidateID
		n.State = Follower

		// IMPORTANT:
		// Granting a vote means this election is legitimate.
		// Reset our election timer so we don't immediately
		// start our own competing election.
		n.LastHeartbeat = time.Now()

		// Randomize the next election timeout.
		n.ElectionTimeout = time.Duration(
			250+rand.Intn(200),
		) * time.Millisecond

		reply.VoteGranted = true

		if err := n.saveStateLocked(); err != nil {
			reply.VoteGranted = false
		}
	}

	return reply
}

func (n *Node) Status() State {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.State
}
func (n *Node) RecoverLogFromPeer(peer string) error {
	n.mu.Lock()
	localLastIndex := n.Log.LastIndex()
	localLastTerm := n.Log.LastTerm()
	n.mu.Unlock()

	args := AppendEntriesArgs{
		Term:         n.CurrentTerm,
		LeaderID:     "",
		PrevLogIndex: localLastIndex,
		PrevLogTerm:  localLastTerm,
		Entries:      nil,
		LeaderCommit: n.CommitIndex,
	}

	reply, err := SendAppendEntries(peer, args)
	if err != nil {
		return err
	}

	if !reply.Success {
		return fmt.Errorf("peer rejected log recovery")
	}

	return nil
}
