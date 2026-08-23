package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/anjuajay2468-glitch/cloudscale/internal/auth"
	"github.com/anjuajay2468-glitch/cloudscale/internal/raft"
	"github.com/anjuajay2468-glitch/cloudscale/internal/replication"
	"github.com/anjuajay2468-glitch/cloudscale/internal/storage"
)

func main() {
	nodeID := flag.String("id", "node1", "unique node ID")
	port := flag.String("port", "8080", "HTTP port")
	dataDir := flag.String("data", "./data", "storage directory")
	peers := flag.String("peers", "", "comma-separated peer URLs")

	flag.Parse()

	// ------------------------------------------------------------
	// Internal node-to-node authentication
	// ------------------------------------------------------------

	internalToken := os.Getenv("CLOUDSCALE_INTERNAL_TOKEN")

	if internalToken == "" {
		log.Fatal(
			"CLOUDSCALE_INTERNAL_TOKEN environment variable is required",
		)
	}

	// ------------------------------------------------------------
	// Storage
	// ------------------------------------------------------------

	store, err := storage.NewStore(*dataDir)
	if err != nil {
		log.Fatal(err)
	}

	// ------------------------------------------------------------
	// Peer configuration
	// ------------------------------------------------------------

	var peerList []string

	if *peers != "" {
		peerList = strings.Split(*peers, ",")
	}
	clusterSize := len(peerList) + 1

	// ------------------------------------------------------------
	// Replication
	// ------------------------------------------------------------

	replicator := replication.NewReplicator(peerList)

	// ------------------------------------------------------------
	// Initialize Raft
	// ------------------------------------------------------------

	raftLogPath := filepath.Join(
		*dataDir,
		"raft-log.json",
	)

	raftStatePath := filepath.Join(
		*dataDir,
		"raft-state.json",
	)

	raftNode, err := raft.NewNodeWithLogPath(
		*nodeID,
		clusterSize,
		raftLogPath,
		raftStatePath,
	)

	if err != nil {
		log.Fatalf(
			"failed to initialize Raft node: %v",
			err,
		)
	}

	// Replay committed Raft entries into the local state machine.
	if err := raftNode.ApplyCommitted(
		func(entry raft.LogEntry) error {
			return applyRaftEntry(store, entry)
		},
	); err != nil {
		log.Fatalf(
			"failed to replay Raft log: %v",
			err,
		)
	}

	// ------------------------------------------------------------
	// Raft HTTP server
	// ------------------------------------------------------------

	raftServer := raft.NewServer(raftNode)

	raftServer.SetInternalToken(
		internalToken,
	)

	raftServer.SetApplyFunc(
		func(entry raft.LogEntry) error {
			return applyRaftEntry(store, entry)
		},
	)

	// Register Raft endpoints.
	raftServer.RegisterRoutes(
		http.DefaultServeMux,
	)

	// ------------------------------------------------------------
	// Start Raft election / heartbeat loop
	// ------------------------------------------------------------

	stopRaft := make(chan struct{})

	go raftNode.RunElectionTimer(
		peerList,
		stopRaft,
	)

	// ------------------------------------------------------------
	// Recover objects from peers
	// ------------------------------------------------------------

	if len(peerList) > 0 {
		fmt.Println(
			"Synchronizing with peers...",
		)

		if err := replicator.SyncFromPeers(store); err != nil {
			log.Printf(
				"initial synchronization failed: %v",
				err,
			)
		}
	}

	// ------------------------------------------------------------
	// Health endpoint
	// ------------------------------------------------------------

	http.HandleFunc(
		"/health",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(
					w,
					"method not allowed",
					http.StatusMethodNotAllowed,
				)
				return
			}

			w.WriteHeader(http.StatusOK)

			fmt.Fprintf(
				w,
				"CloudScale node %s is healthy\n",
				*nodeID,
			)
		},
	)

	// ------------------------------------------------------------
	// Internal object listing endpoint
	// ------------------------------------------------------------

	internalObjectsHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			if r.Method != http.MethodGet {
				http.Error(
					w,
					"method not allowed",
					http.StatusMethodNotAllowed,
				)
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

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			if err := json.NewEncoder(
				w,
			).Encode(objects); err != nil {
				http.Error(
					w,
					"failed to encode objects",
					http.StatusInternalServerError,
				)
			}
		},
	)

	http.Handle(
		"/internal/objects",
		auth.InternalMiddleware(
			internalToken,
			internalObjectsHandler,
		),
	)

	// ------------------------------------------------------------
	// Internal replication / recovery endpoint
	// ------------------------------------------------------------

	internalReplicateHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			name := strings.TrimPrefix(
				r.URL.Path,
				"/internal/replicate/",
			)

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

				if err := store.Put(
					name,
					data,
				); err != nil {
					http.Error(
						w,
						"failed to store replica",
						http.StatusInternalServerError,
					)
					return
				}

				w.WriteHeader(
					http.StatusCreated,
				)

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

				w.WriteHeader(
					http.StatusOK,
				)

				_, _ = w.Write(data)

			default:

				http.Error(
					w,
					"method not allowed",
					http.StatusMethodNotAllowed,
				)
			}
		},
	)

	http.Handle(
		"/internal/replicate/",
		auth.InternalMiddleware(
			internalToken,
			internalReplicateHandler,
		),
	)

	// ------------------------------------------------------------
	// Public object API
	// ------------------------------------------------------------

	http.HandleFunc(
		"/objects/",
		func(w http.ResponseWriter, r *http.Request) {

			name := strings.TrimPrefix(
				r.URL.Path,
				"/objects/",
			)

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

				// Only the Raft leader accepts client writes.
				if raftNode.Status() != raft.Leader {
					http.Error(
						w,
						"node is not the Raft leader",
						http.StatusServiceUnavailable,
					)
					return
				}

				entry, err := raftNode.Propose(
					"PUT",
					name,
					data,
				)

				if err != nil {
					http.Error(
						w,
						err.Error(),
						http.StatusServiceUnavailable,
					)
					return
				}

				committed := raftNode.ReplicateAndCommit(
					peerList,
					entry.Index,
				)

				if !committed {
					http.Error(
						w,
						"failed to achieve Raft majority",
						http.StatusServiceUnavailable,
					)
					return
				}

				if err := raftNode.ApplyCommitted(
					func(entry raft.LogEntry) error {
						return applyRaftEntry(
							store,
							entry,
						)
					},
				); err != nil {
					http.Error(
						w,
						fmt.Sprintf(
							"failed to apply committed entry: %v",
							err,
						),
						http.StatusInternalServerError,
					)
					return
				}

				w.WriteHeader(
					http.StatusCreated,
				)

				fmt.Fprintf(
					w,
					"Object committed through Raft on node %s\n",
					*nodeID,
				)

			case http.MethodGet:

				// Consistent reads are served only by the
				// current Raft leader.
				if raftNode.Status() != raft.Leader {
					http.Error(
						w,
						"node is not the Raft leader",
						http.StatusServiceUnavailable,
					)
					return
				}

				if !raftNode.ConfirmLeadership(
					peerList,
				) {
					http.Error(
						w,
						"unable to confirm Raft leadership",
						http.StatusServiceUnavailable,
					)
					return
				}

				if err := raftNode.ApplyCommitted(
					func(entry raft.LogEntry) error {
						return applyRaftEntry(
							store,
							entry,
						)
					},
				); err != nil {
					http.Error(
						w,
						fmt.Sprintf(
							"failed to apply committed entries: %v",
							err,
						),
						http.StatusInternalServerError,
					)
					return
				}

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

				w.WriteHeader(
					http.StatusOK,
				)

				_, _ = w.Write(data)

			case http.MethodDelete:

				// Only the Raft leader accepts deletes.
				if raftNode.Status() != raft.Leader {
					http.Error(
						w,
						"node is not the Raft leader",
						http.StatusServiceUnavailable,
					)
					return
				}

				entry, err := raftNode.Propose(
					"DELETE",
					name,
					nil,
				)

				if err != nil {
					http.Error(
						w,
						err.Error(),
						http.StatusServiceUnavailable,
					)
					return
				}

				committed := raftNode.ReplicateAndCommit(
					peerList,
					entry.Index,
				)

				if !committed {
					http.Error(
						w,
						"failed to achieve Raft majority",
						http.StatusServiceUnavailable,
					)
					return
				}

				if err := raftNode.ApplyCommitted(
					func(entry raft.LogEntry) error {
						return applyRaftEntry(
							store,
							entry,
						)
					},
				); err != nil {
					http.Error(
						w,
						fmt.Sprintf(
							"failed to apply committed entry: %v",
							err,
						),
						http.StatusInternalServerError,
					)
					return
				}

				w.WriteHeader(
					http.StatusNoContent,
				)

			default:

				http.Error(
					w,
					"method not allowed",
					http.StatusMethodNotAllowed,
				)
			}
		},
	)

	// ------------------------------------------------------------
	// Start HTTP server
	// ------------------------------------------------------------

	address := ":" + *port

	fmt.Printf(
		"CloudScale node %s running on %s\n",
		*nodeID,
		address,
	)

	fmt.Printf(
		"Peers: %s\n",
		strings.Join(peerList, ", "),
	)

	log.Fatal(
		http.ListenAndServe(
			address,
			nil,
		),
	)
}
