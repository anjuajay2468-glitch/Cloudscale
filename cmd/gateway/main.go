package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/anjuajay2468-glitch/cloudscale/internal/auth"
)

type Gateway struct {
	Nodes       []string
	Client      *http.Client
	LeaderCache string
}

type RaftStatus struct {
	ID    string `json:"id"`
	State string `json:"state"`
	Term  int    `json:"term"`
}

func NewGateway(nodes []string) *Gateway {
	return &Gateway{
		Nodes: nodes,
		Client: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

// findLeader checks every configured Raft node and returns
// the node currently reporting itself as Leader.
func (g *Gateway) findLeader() (string, error) {
	for _, node := range g.Nodes {
		node = strings.TrimSpace(node)

		if node == "" {
			continue
		}

		resp, err := g.Client.Get(node + "/raft/status")
		if err != nil {
			continue
		}

		var status RaftStatus

		err = json.NewDecoder(resp.Body).Decode(&status)
		resp.Body.Close()

		if err != nil {
			continue
		}

		if status.State == "Leader" {
			g.LeaderCache = node
			return node, nil
		}
	}

	return "", fmt.Errorf("no Raft leader available")
}

// getLeader first checks the cached leader.
// If the cache is stale, it discovers the leader again.
func (g *Gateway) getLeader() (string, error) {
	if g.LeaderCache != "" {
		resp, err := g.Client.Get(g.LeaderCache + "/raft/status")

		if err == nil {
			var status RaftStatus

			decodeErr := json.NewDecoder(resp.Body).Decode(&status)
			resp.Body.Close()

			if decodeErr == nil && status.State == "Leader" {
				return g.LeaderCache, nil
			}
		}

		g.LeaderCache = ""
	}

	return g.findLeader()
}

// forward sends an authenticated client request to the
// current Raft leader.
func (g *Gateway) forward(w http.ResponseWriter, r *http.Request) {
	leader, err := g.getLeader()
	if err != nil {
		http.Error(
			w,
			"no Raft leader available",
			http.StatusServiceUnavailable,
		)
		return
	}

	targetURL := leader + r.URL.Path

	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	var body io.Reader

	if r.Body != nil {
		body = r.Body
	}

	req, err := http.NewRequest(
		r.Method,
		targetURL,
		body,
	)

	if err != nil {
		http.Error(
			w,
			"failed to create upstream request",
			http.StatusInternalServerError,
		)
		return
	}

	req.Header = r.Header.Clone()

	resp, err := g.Client.Do(req)

	if err != nil {
		g.LeaderCache = ""

		leader, err = g.findLeader()

		if err != nil {
			http.Error(
				w,
				"no Raft leader available",
				http.StatusServiceUnavailable,
			)
			return
		}

		_ = leader

		http.Error(
			w,
			"leader changed; please retry request",
			http.StatusServiceUnavailable,
		)
		return
	}

	defer resp.Body.Close()

	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	w.WriteHeader(resp.StatusCode)

	_, _ = io.Copy(w, resp.Body)
}

func (g *Gateway) handleObjects(w http.ResponseWriter, r *http.Request) {
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
	case http.MethodGet, http.MethodPut, http.MethodDelete:
		g.forward(w, r)

	default:
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
	}
}

func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	leader, err := g.getLeader()

	if err != nil {
		http.Error(
			w,
			"gateway has no available Raft leader",
			http.StatusServiceUnavailable,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"status": "healthy",
		"leader": leader,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(
			w,
			"failed to encode health response",
			http.StatusInternalServerError,
		)
	}
}

func main() {
	port := flag.String(
		"port",
		"9000",
		"gateway HTTP port",
	)

	nodes := flag.String(
		"nodes",
		"http://localhost:8001,http://localhost:8002,http://localhost:8003",
		"comma-separated Raft node URLs",
	)

	flag.Parse()

	nodeList := strings.Split(*nodes, ",")

	adminKey := os.Getenv("CLOUDSCALE_ADMIN_KEY")
	readerKey := os.Getenv("CLOUDSCALE_READER_KEY")

	if adminKey == "" {
		log.Fatal(
			"CLOUDSCALE_ADMIN_KEY environment variable is required",
		)
	}

	if readerKey == "" {
		log.Fatal(
			"CLOUDSCALE_READER_KEY environment variable is required",
		)
	}

	authenticator := auth.NewWithKeys(
		map[string]auth.APIKey{
			"admin": {
				Key:  []byte(adminKey),
				Role: auth.RoleAdmin,
			},
			"reader": {
				Key:  []byte(readerKey),
				Role: auth.RoleReader,
			},
		},
	)

	gateway := NewGateway(nodeList)

	// /health remains public.
	http.HandleFunc(
		"/health",
		gateway.handleHealth,
	)

	// /objects/* requires authentication and authorization.
	protectedObjects := authenticator.Middleware(
		http.HandlerFunc(gateway.handleObjects),
	)

	http.Handle(
		"/objects/",
		protectedObjects,
	)

	address := ":" + *port

	fmt.Printf(
		"CloudScale Gateway running on %s\n",
		address,
	)

	fmt.Printf(
		"Raft nodes: %s\n",
		strings.Join(nodeList, ", "),
	)

	log.Fatal(
		http.ListenAndServe(address, nil),
	)
}
