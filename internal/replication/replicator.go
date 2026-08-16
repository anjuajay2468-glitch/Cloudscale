package replication

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type Replicator struct {
	Peers []string
	Client *http.Client
}

func NewReplicator(peers []string) *Replicator {
	return &Replicator{
		Peers: peers,
		Client: &http.Client{},
	}
}

func (r *Replicator) ReplicatePut(name string, data []byte) error {
	for _, peer := range r.Peers {
		url := fmt.Sprintf("%s/objects/%s", peer, name)

		req, err := http.NewRequest(
			http.MethodPut,
			url,
			bytes.NewReader(data),
		)

		if err != nil {
			return err
		}

		resp, err := r.Client.Do(req)
		if err != nil {
			return fmt.Errorf("replication to %s failed: %w", peer, err)
		}

		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf(
				"replication to %s returned status %d",
				peer,
				resp.StatusCode,
			)
		}
	}

	return nil
}