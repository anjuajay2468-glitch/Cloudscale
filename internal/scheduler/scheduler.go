package scheduler

import "fmt"

// Scheduler chooses the best node for a workload.
type Scheduler struct {
}

// NewScheduler creates a new scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// SelectNode chooses a node capable of running the workload.
//
// Phase 14 initially uses a deterministic baseline algorithm.
// We will replace the scoring function with an AI-assisted
// model later.
func (s *Scheduler) SelectNode(
	workload Workload,
	nodes []NodeResources,
) (NodeResources, error) {

	var best NodeResources
	found := false

	for _, node := range nodes {
		if !node.CanRun(workload) {
			continue
		}

		if !found {
			best = node
			found = true
			continue
		}

		// Baseline policy:
		// prefer the node with more available CPU.
		if node.AvailableCPU() > best.AvailableCPU() {
			best = node
		}
	}

	if !found {
		return NodeResources{},
			fmt.Errorf("no suitable node available for workload %s", workload.ID)
	}

	return best, nil
}
