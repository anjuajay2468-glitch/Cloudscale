package raft

import (
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

	CommitIndex int

	LastApplied int

	NextIndex  map[string]int
	
	MatchIndex map[string]int
}

func NewNode(id string) *Node {
	return &Node{
	ID:              id,
	State:            Follower,
	CurrentTerm:     0,
	VotedFor:        "",
	LastHeartbeat:   time.Now(),
	ClusterSize:     3,
	ElectionTimeout: time.Duration(250+rand.Intn(200)) * time.Millisecond,

	Log:         NewLog(),
	CommitIndex: 0,
	LastApplied: 0,

	NextIndex:  make(map[string]int),
	MatchIndex: make(map[string]int),
}
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

	// Candidate is from an older term.
	if args.Term < n.CurrentTerm {
		return reply
	}

	// Candidate has a newer term.
	if args.Term > n.CurrentTerm {
		n.CurrentTerm = args.Term
		n.State = Follower
		n.VotedFor = ""
	}

	reply.Term = n.CurrentTerm

	// Grant vote if we haven't voted for another candidate
	// in this term.
	if n.VotedFor == "" || n.VotedFor == args.CandidateID {
		n.VotedFor = args.CandidateID
		reply.VoteGranted = true
		n.LastHeartbeat = time.Now()
	}

	return reply
}