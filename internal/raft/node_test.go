package raft

import (
	"testing"
	"time"
)

func TestNewNode(t *testing.T) {
	node := NewNode("node1", 3)

	if node.ID != "node1" {
		t.Fatalf("expected node1, got %s", node.ID)
	}

	if node.State != Follower {
		t.Fatalf("expected follower state")
	}

	if node.CurrentTerm != 0 {
		t.Fatalf("expected initial term 0")
	}

	if node.VotedFor != "" {
		t.Fatalf("expected no initial vote")
	}
}
func TestRequestVote(t *testing.T) {
	node := NewNode("node1", 3)

	reply := node.RequestVote(RequestVoteArgs{
		Term:        1,
		CandidateID: "node2",
	})

	if !reply.VoteGranted {
		t.Fatal("expected vote to be granted")
	}

	if node.VotedFor != "node2" {
		t.Fatalf("expected vote for node2, got %s", node.VotedFor)
	}

	// Node 1 should not vote for another candidate
	// in the same term.
	reply = node.RequestVote(RequestVoteArgs{
		Term:        1,
		CandidateID: "node3",
	})

	if reply.VoteGranted {
		t.Fatal("node should not grant a second vote in the same term")
	}
}
func TestElectionWithoutPeersFails(t *testing.T) {
	node := NewNode("node1", 3)

	elected := node.StartElection([]string{})

	if elected {
		t.Fatal("node should not become leader without a majority")
	}

	if node.State != Candidate {
		t.Fatalf("expected candidate state, got %s", node.State)
	}
}
func TestElectionMajority(t *testing.T) {
	node := NewNode("node1", 3)

	// We don't have real peers here, so this should not
	// become leader.
	elected := node.StartElection([]string{
		"http://localhost:9998",
		"http://localhost:9999",
	})

	if elected {
		t.Fatal("election should fail when peers are unavailable")
	}

	if node.State != Candidate {
		t.Fatalf("expected candidate state, got %s", node.State)
	}

	if node.CurrentTerm != 1 {
		t.Fatalf("expected term 1, got %d", node.CurrentTerm)
	}
}
func TestAppendEntriesHeartbeat(t *testing.T) {
	node := NewNode("node2", 3)

	node.CurrentTerm = 1
	node.State = Candidate

	reply := node.HandleAppendEntries(AppendEntriesArgs{
		Term:     1,
		LeaderID: "node1",
	})

	if !reply.Success {
		t.Fatal("expected heartbeat to be accepted")
	}

	if node.State != Follower {
		t.Fatalf("expected follower, got %s", node.State)
	}

	if node.CurrentTerm != 1 {
		t.Fatalf("expected term 1, got %d", node.CurrentTerm)
	}
}
func TestHeartbeatMakesCandidateFollower(t *testing.T) {
	node := NewNode("node2", 3)

	node.CurrentTerm = 1
	node.State = Candidate

	reply := node.HandleAppendEntries(AppendEntriesArgs{
		Term:     1,
		LeaderID: "node1",
	})

	if !reply.Success {
		t.Fatal("expected heartbeat to succeed")
	}

	if node.State != Follower {
		t.Fatalf(
			"expected follower after heartbeat, got %s",
			node.State,
		)
	}
}
func TestElectionTimerExpires(t *testing.T) {
	node := NewNode("node1", 3)

	node.ElectionTimeout = 10 * time.Millisecond
	node.LastHeartbeat = time.Now().Add(-20 * time.Millisecond)

	if !node.ElectionTimerExpired() {
		t.Fatal("expected election timer to expire")
	}
}
func TestLeaderElectionTimerDoesNotExpire(t *testing.T) {
	node := NewNode("node1", 3)

	node.State = Leader
	node.ElectionTimeout = 10 * time.Millisecond
	node.LastHeartbeat = time.Now().Add(-20 * time.Millisecond)

	if node.ElectionTimerExpired() {
		t.Fatal("leader election timer should not expire")
	}
}
func TestAppendEntriesReplicatesLog(t *testing.T) {
	node := NewNode("node2", 3)

	node.CurrentTerm = 1

	entry := LogEntry{
		Index:   1,
		Term:    1,
		Command: "PUT",
		Key:     "hello.txt",
		Data:    []byte("Hello CloudScale"),
	}

	reply := node.HandleAppendEntries(AppendEntriesArgs{
		Term:         1,
		LeaderID:     "node1",
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		Entries:      []LogEntry{entry},
		LeaderCommit: 0,
	})

	if !reply.Success {
		t.Fatal("expected append entries to succeed")
	}

	if len(node.Log.Entries) != 1 {
		t.Fatalf(
			"expected 1 log entry, got %d",
			len(node.Log.Entries),
		)
	}

	if node.Log.Entries[0].Key != "hello.txt" {
		t.Fatalf(
			"expected hello.txt, got %s",
			node.Log.Entries[0].Key,
		)
	}
}
func TestAppendEntriesResolvesConflict(t *testing.T) {
	node := NewNode("node2", 3)

	node.CurrentTerm = 2

	// Existing follower log.
	node.Log.Append(LogEntry{
		Index:   1,
		Term:    1,
		Command: "PUT",
		Key:     "a.txt",
		Data:    []byte("A"),
	})

	node.Log.Append(LogEntry{
		Index:   2,
		Term:    1,
		Command: "PUT",
		Key:     "b.txt",
		Data:    []byte("B"),
	})

	// Conflicting entry.
	node.Log.Append(LogEntry{
		Index:   3,
		Term:    9,
		Command: "PUT",
		Key:     "wrong.txt",
		Data:    []byte("WRONG"),
	})

	// Leader says entry 3 should actually be term 2.
	reply := node.HandleAppendEntries(AppendEntriesArgs{
		Term:         2,
		LeaderID:     "node1",
		PrevLogIndex: 2,
		PrevLogTerm:  1,
		Entries: []LogEntry{
			{
				Index:   3,
				Term:    2,
				Command: "PUT",
				Key:     "correct.txt",
				Data:    []byte("CORRECT"),
			},
		},
		LeaderCommit: 0,
	})

	if !reply.Success {
		t.Fatal("expected append entries to succeed")
	}

	if len(node.Log.Entries) != 3 {
		t.Fatalf(
			"expected 3 log entries, got %d",
			len(node.Log.Entries),
		)
	}

	if node.Log.Entries[2].Term != 2 {
		t.Fatalf(
			"expected term 2, got %d",
			node.Log.Entries[2].Term,
		)
	}

	if node.Log.Entries[2].Key != "correct.txt" {
		t.Fatalf(
			"expected correct.txt, got %s",
			node.Log.Entries[2].Key,
		)
	}
}
func TestInitializeLeaderReplication(t *testing.T) {
	node := NewNode("node1", 3)

	node.Log.Append(LogEntry{
		Index:   1,
		Term:    1,
		Command: "PUT",
		Key:     "a.txt",
		Data:    []byte("A"),
	})

	node.Log.Append(LogEntry{
		Index:   2,
		Term:    1,
		Command: "PUT",
		Key:     "b.txt",
		Data:    []byte("B"),
	})

	peers := []string{
		"http://localhost:8002",
		"http://localhost:8003",
	}

	node.InitializeLeaderReplication(peers)

	if node.NextIndex[peers[0]] != 3 {
		t.Fatalf(
			"expected nextIndex 3, got %d",
			node.NextIndex[peers[0]],
		)
	}

	if node.NextIndex[peers[1]] != 3 {
		t.Fatalf(
			"expected nextIndex 3, got %d",
			node.NextIndex[peers[1]],
		)
	}

	if node.MatchIndex[peers[0]] != 0 {
		t.Fatalf(
			"expected matchIndex 0, got %d",
			node.MatchIndex[peers[0]],
		)
	}
}
func TestAdvanceCommitIndex(t *testing.T) {
	node := NewNode("node1", 3)

	node.State = Leader

	node.Log.Append(LogEntry{
		Index:   1,
		Term:    1,
		Command: "PUT",
		Key:     "a.txt",
		Data:    []byte("A"),
	})

	node.Log.Append(LogEntry{
		Index:   2,
		Term:    1,
		Command: "PUT",
		Key:     "b.txt",
		Data:    []byte("B"),
	})

	node.MatchIndex["node2"] = 2
	node.MatchIndex["node3"] = 1

	node.AdvanceCommitIndex()

	if node.CommitIndex != 2 {
		t.Fatalf(
			"expected commit index 2, got %d",
			node.CommitIndex,
		)
	}
}
func TestApplyCommitted(t *testing.T) {
	node := NewNode("node1", 3)

	node.Log.Append(LogEntry{
		Index:   1,
		Term:    1,
		Command: "PUT",
		Key:     "hello.txt",
		Data:    []byte("Hello CloudScale"),
	})

	node.Log.Append(LogEntry{
		Index:   2,
		Term:    1,
		Command: "DELETE",
		Key:     "old.txt",
	})

	node.CommitIndex = 2

	var applied []LogEntry

	err := node.ApplyCommitted(func(entry LogEntry) error {
		applied = append(applied, entry)
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(applied) != 2 {
		t.Fatalf(
			"expected 2 applied entries, got %d",
			len(applied),
		)
	}

	if applied[0].Key != "hello.txt" {
		t.Fatalf(
			"expected hello.txt, got %s",
			applied[0].Key,
		)
	}

	if applied[1].Command != "DELETE" {
		t.Fatalf(
			"expected DELETE, got %s",
			applied[1].Command,
		)
	}

	if node.LastApplied != 2 {
		t.Fatalf(
			"expected LastApplied 2, got %d",
			node.LastApplied,
		)
	}
}
func TestProposeCreatesLogEntry(t *testing.T) {
	node := NewNode("node1", 3)

	node.State = Leader
	node.CurrentTerm = 7

	entry, err := node.Propose(
		"PUT",
		"hello.txt",
		[]byte("Hello CloudScale"),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.Index != 1 {
		t.Fatalf("expected index 1, got %d", entry.Index)
	}

	if entry.Term != 7 {
		t.Fatalf("expected term 7, got %d", entry.Term)
	}

	if entry.Command != "PUT" {
		t.Fatalf("expected PUT, got %s", entry.Command)
	}

	if entry.Key != "hello.txt" {
		t.Fatalf("expected hello.txt, got %s", entry.Key)
	}

	if string(entry.Data) != "Hello CloudScale" {
		t.Fatalf(
			"expected Hello CloudScale, got %s",
			string(entry.Data),
		)
	}

	if node.Log.LastIndex() != 1 {
		t.Fatalf(
			"expected log index 1, got %d",
			node.Log.LastIndex(),
		)
	}
}
func TestFollowerCannotPropose(t *testing.T) {
	node := NewNode("node2", 3)

	node.State = Follower

	_, err := node.Propose(
		"PUT",
		"hello.txt",
		[]byte("Hello"),
	)

	if err == nil {
		t.Fatal("expected follower proposal to fail")
	}

	if node.Log.LastIndex() != 0 {
		t.Fatalf(
			"expected empty log, got index %d",
			node.Log.LastIndex(),
		)
	}
}
func TestLeaderReplicationTracking(t *testing.T) {
	node := NewNode("node1", 3)

	node.State = Leader
	node.CurrentTerm = 1

	node.Log.Append(LogEntry{
		Index:   1,
		Term:    1,
		Command: "PUT",
		Key:     "hello.txt",
		Data:    []byte("Hello"),
	})

	peers := []string{
		"http://localhost:8002",
		"http://localhost:8003",
	}

	node.InitializeLeaderReplication(peers)

	if node.NextIndex[peers[0]] != 2 {
		t.Fatalf(
			"expected nextIndex 2, got %d",
			node.NextIndex[peers[0]],
		)
	}

	if node.MatchIndex[peers[0]] != 0 {
		t.Fatalf(
			"expected matchIndex 0, got %d",
			node.MatchIndex[peers[0]],
		)
	}
}
func TestMajorityCommitWithThreeNodes(t *testing.T) {
	node := NewNode("node1", 3)

	node.State = Leader
	node.CurrentTerm = 1

	node.Log.Append(LogEntry{
		Index:   1,
		Term:    1,
		Command: "PUT",
		Key:     "hello.txt",
		Data:    []byte("Hello CloudScale"),
	})

	node.MatchIndex["node2"] = 1
	node.MatchIndex["node3"] = 0

	node.AdvanceCommitIndex()

	if node.CommitIndex != 1 {
		t.Fatalf(
			"expected commit index 1, got %d",
			node.CommitIndex,
		)
	}
}
func TestReplicatedEntryCanBeCommitted(t *testing.T) {
	node := NewNode("node1", 3)

	node.State = Leader
	node.CurrentTerm = 1

	node.Log.Append(LogEntry{
		Index:   1,
		Term:    1,
		Command: "PUT",
		Key:     "hello.txt",
		Data:    []byte("Hello CloudScale"),
	})

	node.MatchIndex["node2"] = 1
	node.MatchIndex["node3"] = 0

	node.AdvanceCommitIndex()

	if node.CommitIndex != 1 {
		t.Fatalf(
			"expected commit index 1, got %d",
			node.CommitIndex,
		)
	}
}
func TestLogPersistence(t *testing.T) {
	path := t.TempDir() + "/raft-log.json"

	log1 := NewLog()

	log1.Append(LogEntry{
		Index:   1,
		Term:    5,
		Command: "PUT",
		Key:     "hello.txt",
		Data:    []byte("Hello CloudScale"),
	})

	log1.Append(LogEntry{
		Index:   2,
		Term:    5,
		Command: "DELETE",
		Key:     "old.txt",
	})

	if err := log1.Save(path); err != nil {
		t.Fatalf("failed to save log: %v", err)
	}

	log2 := NewLog()

	if err := log2.Load(path); err != nil {
		t.Fatalf("failed to load log: %v", err)
	}

	if len(log2.Entries) != 2 {
		t.Fatalf(
			"expected 2 entries, got %d",
			len(log2.Entries),
		)
	}

	if log2.Entries[0].Command != "PUT" {
		t.Fatalf("expected PUT entry")
	}

	if log2.Entries[0].Key != "hello.txt" {
		t.Fatalf("expected hello.txt")
	}

	if string(log2.Entries[0].Data) != "Hello CloudScale" {
		t.Fatalf("unexpected entry data")
	}

	if log2.Entries[1].Command != "DELETE" {
		t.Fatalf("expected DELETE entry")
	}
}
