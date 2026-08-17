package main

import (
	"testing"

	"github.com/anjuajay2468-glitch/cloudscale/internal/raft"
	"github.com/anjuajay2468-glitch/cloudscale/internal/storage"
)

func TestApplyRaftPut(t *testing.T) {
	store, err := storage.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	entry := raft.LogEntry{
		Index:   1,
		Term:    1,
		Command: "PUT",
		Key:     "hello.txt",
		Data:    []byte("Hello CloudScale"),
	}

	if err := applyRaftEntry(store, entry); err != nil {
		t.Fatalf("applyRaftEntry failed: %v", err)
	}

	data, err := store.Get("hello.txt")
	if err != nil {
		t.Fatalf("failed to read object: %v", err)
	}

	if string(data) != "Hello CloudScale" {
		t.Fatalf(
			"expected Hello CloudScale, got %s",
			string(data),
		)
	}
}

func TestApplyRaftDelete(t *testing.T) {
	store, err := storage.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	if err := store.Put(
		"delete-me.txt",
		[]byte("temporary"),
	); err != nil {
		t.Fatalf("failed to create test object: %v", err)
	}

	entry := raft.LogEntry{
		Index:   1,
		Term:    1,
		Command: "DELETE",
		Key:     "delete-me.txt",
	}

	if err := applyRaftEntry(store, entry); err != nil {
		t.Fatalf("applyRaftEntry failed: %v", err)
	}

	_, err = store.Get("delete-me.txt")

	if err != storage.ErrNotFound {
		t.Fatalf("expected object to be deleted")
	}
}
