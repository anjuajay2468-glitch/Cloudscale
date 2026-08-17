package raft

import "time"

func (n *Node) RunElectionTimer(
	peers []string,
	stop <-chan struct{},
) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			n.mu.Lock()
			state := n.State
			n.mu.Unlock()

			// Leaders send heartbeats instead of starting elections.
			if state == Leader {
				n.SendHeartbeats(peers)
				continue
			}

			// Followers and candidates start an election
			// when their timeout expires.
			if n.ElectionTimerExpired() {
				elected := n.StartElection(peers)

				if elected {
					// Immediately notify followers that we are
					// the new leader.
					n.SendHeartbeats(peers)
				}
			}

		case <-stop:
			return
		}
	}
}