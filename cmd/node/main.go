package main

import (
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

	peers := flag.String(
		"peers",
		"",
		"comma-separated peer URLs",
	)

	flag.Parse()

	var peerList []string
	if *peers != "" {
		peerList = strings.Split(*peers, ",")
	}
	replicator := replication.NewReplicator(peerList)

	store, err := storage.NewStore(*dataDir)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "CloudScale node %s is healthy\n", *nodeID)
	})

	http.HandleFunc("/internal/replicate/", func(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Path[len("/internal/replicate/"):]

	if name == "" {
		http.Error(w, "object name is required", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	if err := store.Put(name, data); err != nil {
		http.Error(w, "failed to store replica", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Replica stored on node %s\n", *nodeID)
})

	http.HandleFunc("/objects/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len("/objects/"):]

		if name == "" {
			http.Error(w, "object name is required", http.StatusBadRequest)
			return
		}

		switch r.Method {

		case http.MethodPut:
			data, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusBadRequest)
				return
			}

			err = store.Put(name, data)
			if err != nil {
				http.Error(w, "failed to store object", http.StatusInternalServerError)
				return
			}

			if err := replicator.ReplicatePut(name, data); err != nil {
				http.Error(
					w,
					fmt.Sprintf("replication failed: %v", err),
					http.StatusInternalServerError,
				)
				return
			}

			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, "Object stored on node %s\n", *nodeID)

		case http.MethodGet:
			data, err := store.Get(name)

			if err == storage.ErrNotFound {
				http.Error(w, "object not found", http.StatusNotFound)
				return
			}

			if err != nil {
				http.Error(w, "failed to read object", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)

		case http.MethodDelete:
			err := store.Delete(name)

			if err == storage.ErrNotFound {
				http.Error(w, "object not found", http.StatusNotFound)
				return
			}

			if err != nil {
				http.Error(w, "failed to delete object", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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