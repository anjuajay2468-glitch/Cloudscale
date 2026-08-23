package raft

import (
	"encoding/json"
	"fmt"
	"os"
)

type LogEntry struct {
	Index   int
	Term    int
	Command string
	Key     string
	Data    []byte
}

type Log struct {
	Entries []LogEntry
}

func NewLog() *Log {
	return &Log{
		Entries: make([]LogEntry, 0),
	}
}

func (l *Log) LastIndex() int {
	if len(l.Entries) == 0 {
		return 0
	}

	return l.Entries[len(l.Entries)-1].Index
}

func (l *Log) LastTerm() int {
	if len(l.Entries) == 0 {
		return 0
	}

	return l.Entries[len(l.Entries)-1].Term
}

func (l *Log) Append(entry LogEntry) {
	l.Entries = append(l.Entries, entry)
}

func (l *Log) Save(path string) error {
	data, err := json.MarshalIndent(l.Entries, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write raft log: %w", err)
	}

	return nil
}

func (l *Log) Load(path string) error {
	data, err := os.ReadFile(path)

	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to read raft log: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	var entries []LogEntry

	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to decode raft log: %w", err)
	}

	l.Entries = entries

	return nil
}
