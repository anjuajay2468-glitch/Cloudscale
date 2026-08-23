package scheduler

import (
	"fmt"
	"sync"
)

type LearningScheduler struct {
	mu sync.Mutex

	CPUWeight    float64
	MemoryWeight float64
}

func NewLearningScheduler() *LearningScheduler {
	return &LearningScheduler{
		CPUWeight:    0.5,
		MemoryWeight: 0.5,
	}
}

func (l *LearningScheduler) SelectNode(
	nodes []NodeResources,
	history []WorkloadHistory,
) (NodeResources, float64, error) {

	l.mu.Lock()
	cpuWeight := l.CPUWeight
	memoryWeight := l.MemoryWeight
	l.mu.Unlock()

	var best NodeResources
	bestScore := -1.0
	found := false

	for _, node := range nodes {
		if !node.Healthy {
			continue
		}

		cpuAvailability := 1 - utilization(
			node.CPUUsedMillis,
			node.CPUTotalMillis,
		)

		memoryAvailability := 1 - utilization(
			node.MemoryUsedMB,
			node.MemoryTotalMB,
		)

		historyPenalty := historicalPenalty(
			node.NodeID,
			history,
		)

		score :=
			cpuWeight*cpuAvailability +
				memoryWeight*memoryAvailability -
				0.2*historyPenalty

		if !found || score > bestScore {
			best = node
			bestScore = score
			found = true
		}
	}

	if !found {
		return NodeResources{}, 0, fmt.Errorf(
			"no healthy node available",
		)
	}

	return best, bestScore, nil
}

// Learn updates the CPU/memory weighting using observed error.
//
// If CPU prediction error is larger, CPU gets more weight.
// If memory prediction error is larger, memory gets more weight.
func (l *LearningScheduler) Learn(
	feedback SchedulingFeedback,
) {

	l.mu.Lock()
	defer l.mu.Unlock()

	cpuError := feedback.CPUErrorPercent
	memoryError := feedback.MemoryErrorPercent

	total := cpuError + memoryError

	if total == 0 {
		return
	}

	targetCPUWeight := cpuError / total
	targetMemoryWeight := memoryError / total

	learningRate := 0.1

	l.CPUWeight =
		(1-learningRate)*l.CPUWeight +
			learningRate*targetCPUWeight

	l.MemoryWeight =
		(1-learningRate)*l.MemoryWeight +
			learningRate*targetMemoryWeight

	// Keep the weights normalized.
	totalWeight := l.CPUWeight + l.MemoryWeight

	if totalWeight > 0 {
		l.CPUWeight /= totalWeight
		l.MemoryWeight /= totalWeight
	}
}

func (l *LearningScheduler) Weights() (float64, float64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.CPUWeight, l.MemoryWeight
}
