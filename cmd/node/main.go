package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/anjuajay2468-glitch/cloudscale/internal/replication"
	"github.com/anjuajay2468-glitch/cloudscale/internal/storage"
)

func main() {
	nodeID := flag.String("id", "node1", "unique node ID")
	port := flag.String("port", "8080", "HTTP port")
	dataDir := flag.String("data", "./data", "storage directory")
	peers := flag.String("peers", "", "comma-separated peer URLs")

	flag.Parse()

	store, err := storage.NewStore(*dataDir)
	if err != nil {
		log.Fatal(err)
	}

	var peerList []string

	if *peers != "" {
		peerList = strings.Split(*peers, ",")
	}

	replicator := replication.NewReplicator(peerList)

	// Recover objects from peers when this node starts.
	if len(peerList) > 0 {
		fmt.Println("Synchronizing with peers...")

		if err := replicator.SyncFromPeers(store); err != nil {
			log.Printf("initial synchronization failed: %v", err)
		}
	}

	// Health endpoint.
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "CloudScale node %s is healthy\n", *nodeID)
	})

	// Internal endpoint used to list objects on this node.
	http.HandleFunc("/internal/objects", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		objects, err := store.List()
		if err != nil {
			http.Error(
				w,
				"failed to list objects",
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(objects); err != nil {
			http.Error(
				w,
				"failed to encode objects",
				http.StatusInternalServerError,
			)
		}
	})

	// Internal endpoint used for node-to-node replication and recovery.
	http.HandleFunc("/internal/replicate/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/internal/replicate/")

		if name == "" {
			http.Error(
				w,
				"object name is required",
				http.StatusBadRequest,
			)
			return
		}

		switch r.Method {

		case http.MethodPut:
			data, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(
					w,
					"failed to read request body",
					http.StatusBadRequest,
				)
				return
			}

			if err := store.Put(name, data); err != nil {
				http.Error(
					w,
					"failed to store replica",
					http.StatusInternalServerError,
				)
				return
			}

			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(
				w,
				"Replica stored on node %s\n",
				*nodeID,
			)

		case http.MethodGet:
			data, err := store.Get(name)

			if err == storage.ErrNotFound {
				http.Error(
					w,
					"object not found",
					http.StatusNotFound,
				)
				return
			}

			if err != nil {
				http.Error(
					w,
					"failed to read replica",
					http.StatusInternalServerError,
				)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)

		default:
			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	})

	// Public object API.
	http.HandleFunc("/objects/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/objects/")

		if name == "" {
			http.Error(
				w,
				"object name is required",
				http.StatusBadRequest,
			)
			return
		}

		switch r.Method {

		case http.MethodPut:
			data, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(
					w,
					"failed to read request body",
					http.StatusBadRequest,
				)
				return
			}

			// Store locally first.
			if err := store.Put(name, data); err != nil {
				http.Error(
					w,
					"failed to store object",
					http.StatusInternalServerError,
				)
				return
			}

			// Replicate to peers and enforce write quorum.
			if err := replicator.ReplicatePut(name, data); err != nil {
				http.Error(
					w,
					fmt.Sprintf("replication failed: %v", err),
					http.StatusInternalServerError,
				)
				return
			}

			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(
				w,
				"Object stored on node %s\n",
				*nodeID,
			)

		case http.MethodGet:
			data, err := store.Get(name)

			if err == storage.ErrNotFound {
				http.Error(
					w,
					"object not found",
					http.StatusNotFound,
				)
				return
			}

			if err != nil {
				http.Error(
					w,
					"failed to read object",
					http.StatusInternalServerError,
				)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)

		case http.MethodDelete:
			err := store.Delete(name)

			if err == storage.ErrNotFound {
				http.Error(
					w,
					"object not found",
					http.StatusNotFound,
				)
				return
			}

			if err != nil {
				http.Error(
					w,
					"failed to delete object",
					http.StatusInternalServerError,
				)
				return
			}

			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	})

	address := ":" + *port

	fmt.Printf(
		"CloudScale node %s running on %s\n",
		*nodeID,
		address,
	)

	fmt.Printf(
		"Storage directory: %s\n",
		*dataDir,
	)

	log.Fatal(http.ListenAndServe(address, nil))
}