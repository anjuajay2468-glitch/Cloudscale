package main

import (
	"fmt"

	"github.com/anjuajay2468-glitch/cloudscale/internal/raft"
	"github.com/anjuajay2468-glitch/cloudscale/internal/storage"
)

func applyRaftEntry(store *storage.Store, entry raft.LogEntry) error {
	switch entry.Command {

	case "PUT":
		return store.Put(entry.Key, entry.Data)

	case "DELETE":
		err := store.Delete(entry.Key)

		if err == storage.ErrNotFound {
			// DELETE is already satisfied if the object does not exist.
			return nil
		}

		return err

	default:
		return fmt.Errorf("unknown Raft command: %s", entry.Command)
	}
}
