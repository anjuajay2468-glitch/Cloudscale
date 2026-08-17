package replication

import (
	"bytes"
	"encoding/json"
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
			fmt.Printf("Replication to %s failed: %v\n", peer, err)
			continue
		}

		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			successes++
			fmt.Printf("Replicated %s to %s\n", name, peer)
		} else {
			fmt.Printf(
				"Replication to %s returned status %d\n",
				peer,
				resp.StatusCode,
			)
		}
	}

	if successes < required {
		return fmt.Errorf(
			"write quorum not reached: %d/%d acknowledgements",
			successes,
			required,
		)
	}

	return nil
}
type Store interface {
	List() ([]string, error)
	Put(name string, data []byte) error
}
func (r *Replicator) SyncFromPeers(store Store) error {
	localObjects, err := store.List()
	if err != nil {
		return err
	}

	localSet := make(map[string]bool)

	for _, object := range localObjects {
		localSet[object] = true
	}

	for _, peer := range r.Peers {
		objects, err := r.listPeerObjects(peer)
		if err != nil {
			continue
		}

		for _, object := range objects {
			if localSet[object] {
				continue
			}

			data, err := r.fetchObject(peer, object)
			if err != nil {
				continue
			}

			if err := store.Put(object, data); err != nil {
				return err
			}

			localSet[object] = true

			fmt.Printf(
				"Recovered object %s from %s\n",
				object,
				peer,
			)
		}
	}

	return nil
}

func (r *Replicator) listPeerObjects(peer string) ([]string, error) {
	url := fmt.Sprintf("%s/internal/objects", peer)

	resp, err := r.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"peer returned status %d",
			resp.StatusCode,
		)
	}

	var objects []string

	if err := json.NewDecoder(resp.Body).Decode(&objects); err != nil {
		return nil, err
	}

	return objects, nil
}

func (r *Replicator) fetchObject(peer string, name string) ([]byte, error) {
	url := fmt.Sprintf("%s/internal/replicate/%s", peer, name)

	resp, err := r.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"peer returned status %d",
			resp.StatusCode,
		)
	}

	return io.ReadAll(resp.Body)
}