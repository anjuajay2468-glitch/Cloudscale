package raft

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
