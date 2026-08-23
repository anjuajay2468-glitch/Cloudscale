package scheduler

import "fmt"

type IntelligentScorer struct {
	CPUWeight     float64
	MemoryWeight  float64
	HistoryWeight float64
}

func NewIntelligentScorer() *IntelligentScorer {
	return &IntelligentScorer{
		CPUWeight:     0.40,
		MemoryWeight:  0.40,
		HistoryWeight: 0.20,
	}
}

func (s *IntelligentScorer) Score(
	node NodeResources,
	history []WorkloadHistory,
) float64 {

	if !node.Healthy {
		return -1
	}

	cpuUtilization := utilization(
		node.CPUUsedMillis,
		node.CPUTotalMillis,
	)

	memoryUtilization := utilization(
		node.MemoryUsedMB,
		node.MemoryTotalMB,
	)

	cpuAvailability := 1 - cpuUtilization
	memoryAvailability := 1 - memoryUtilization

	historyPenalty := historicalPenalty(
		node.NodeID,
		history,
	)

	score :=
		s.CPUWeight*cpuAvailability +
			s.MemoryWeight*memoryAvailability +
			s.HistoryWeight*(1-historyPenalty)

	return score
}

func (s *IntelligentScorer) SelectNode(
	nodes []NodeResources,
	history []WorkloadHistory,
) (NodeResources, float64, error) {

	var best NodeResources
	bestScore := -1.0
	found := false

	for _, node := range nodes {
		score := s.Score(node, history)

		if score < 0 {
			continue
		}

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

func utilization(used, total int64) float64 {
	if total <= 0 {
		return 1
	}

	value := float64(used) / float64(total)

	if value < 0 {
		return 0
	}

	if value > 1 {
		return 1
	}

	return value
}

func historicalPenalty(
	nodeID string,
	history []WorkloadHistory,
) float64 {

	var total float64
	var count int

	for _, record := range history {
		if record.NodeID != nodeID {
			continue
		}

		errorValue :=
			(record.CPUErrorPercent +
				record.MemoryErrorPercent) / 200

		total += errorValue
		count++
	}

	if count == 0 {
		return 0
	}

	penalty := total / float64(count)

	if penalty < 0 {
		return 0
	}

	if penalty > 1 {
		return 1
	}

	return penalty
}
