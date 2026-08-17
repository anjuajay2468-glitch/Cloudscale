package raft

import (
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
}

func NewNode(id string) *Node {
	return &Node{
		ID:             id,
		State:          Follower,
		CurrentTerm:    0,
		VotedFor:       "",
		LastHeartbeat:  time.Now(),
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