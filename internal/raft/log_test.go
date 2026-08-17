package raft

import "testing"

func TestLogAppend(t *testing.T) {
	log := NewLog()

	log.Append(LogEntry{
		Index:   1,
		Term:    1,
		Command: "PUT",
		Key:     "hello.txt",
		Data:    []byte("Hello CloudScale"),
	})

	if log.LastIndex() != 1 {
		t.Fatalf(
			"expected last index 1, got %d",
			log.LastIndex(),
		)
	}

	if log.LastTerm() != 1 {
		t.Fatalf(
			"expected last term 1, got %d",
			log.LastTerm(),
		)
	}

	if len(log.Entries) != 1 {
		t.Fatalf(
			"expected 1 log entry, got %d",
			len(log.Entries),
		)
	}
}
func TestLogMultipleEntries(t *testing.T) {
	log := NewLog()

	log.Append(LogEntry{
		Index:   1,
		Term:    1,
		Command: "PUT",
		Key:     "a.txt",
		Data:    []byte("A"),
	})

	log.Append(LogEntry{
		Index:   2,
		Term:    1,
		Command: "PUT",
		Key:     "b.txt",
		Data:    []byte("B"),
	})

	log.Append(LogEntry{
		Index:   3,
		Term:    2,
		Command: "DELETE",
		Key:     "a.txt",
	})

	if log.LastIndex() != 3 {
		t.Fatalf(
			"expected last index 3, got %d",
			log.LastIndex(),
		)
	}

	if log.LastTerm() != 2 {
		t.Fatalf(
			"expected last term 2, got %d",
			log.LastTerm(),
		)
	}
}