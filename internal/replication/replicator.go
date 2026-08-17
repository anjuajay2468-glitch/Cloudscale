package replication

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type Replicator struct {
	Peers      []string
	Client     *http.Client
	WriteQuorum int
}

func NewReplicator(peers []string) *Replicator {
	return &Replicator{
		Peers:       peers,
		Client:      &http.Client{},
		WriteQuorum: 2,
	}
}

func (r *Replicator) ReplicatePut(name string, data []byte) error {
	successes := 1 // local node already stored the object
	required := r.WriteQuorum

	for _, peer := range r.Peers {
		url := fmt.Sprintf("%s/internal/replicate/%s", peer, name)

		req, err := http.NewRequest(
			http.MethodPut,
			url,
			bytes.NewReader(data),
		)

		if err != nil {
			continue
		}

		resp, err := r.Client.Do(req)
		if err != nil {
			continue
		}

		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			successes++
		}

		if successes >= required {
			return nil
		}
	}

	return fmt.Errorf(
		"write quorum not reached: %d/%d acknowledgements",
		successes,
		required,
	)
}