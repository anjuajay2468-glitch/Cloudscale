package scheduler

import (
	"fmt"
	"time"
)

type SchedulerComparison struct {
	WorkloadID string

	BaselineNode    string
	IntelligentNode string

	BaselineScore    float64
	IntelligentScore float64

	BaselineCPUError    float64
	IntelligentCPUError float64

	BaselineMemoryError    float64
	IntelligentMemoryError float64

	RecordedAt time.Time
}

func CompareSchedulers(
	workload Workload,
	nodes []NodeResources,
	history []WorkloadHistory,
	scorer *IntelligentScorer,
) (SchedulerComparison, error) {

	if len(nodes) == 0 {
		return SchedulerComparison{}, fmt.Errorf(
			"no nodes available",
		)
	}

	var baseline NodeResources
	found := false

	for _, node := range nodes {
		if node.Healthy {
			baseline = node
			found = true
			break
		}
	}

	if !found {
		return SchedulerComparison{}, fmt.Errorf(
			"no healthy node available",
		)
	}

	intelligent, intelligentScore, err :=
		scorer.SelectNode(nodes, history)

	if err != nil {
		return SchedulerComparison{}, err
	}

	// Normalize the baseline score to the same 0-1 scale
	// used by the intelligent scheduler.
	baselineScore := baselineScore(baseline)

	baselineCPUError := percentageError(
		workload.CPURequestMillis,
		baseline.CPUUsedMillis,
	)

	baselineMemoryError := percentageError(
		workload.MemoryRequestMB,
		baseline.MemoryUsedMB,
	)

	intelligentCPUError := percentageError(
		workload.CPURequestMillis,
		intelligent.CPUUsedMillis,
	)

	intelligentMemoryError := percentageError(
		workload.MemoryRequestMB,
		intelligent.MemoryUsedMB,
	)

	return SchedulerComparison{
		WorkloadID: workload.ID,

		BaselineNode:    baseline.NodeID,
		IntelligentNode: intelligent.NodeID,

		BaselineScore:    baselineScore,
		IntelligentScore: intelligentScore,

		BaselineCPUError:    baselineCPUError,
		IntelligentCPUError: intelligentCPUError,

		BaselineMemoryError:    baselineMemoryError,
		IntelligentMemoryError: intelligentMemoryError,

		RecordedAt: time.Now(),
	}, nil
}

func baselineScore(node NodeResources) float64 {
	if !node.Healthy {
		return -1
	}

	cpuAvailability := 1 - utilization(
		node.CPUUsedMillis,
		node.CPUTotalMillis,
	)

	memoryAvailability := 1 - utilization(
		node.MemoryUsedMB,
		node.MemoryTotalMB,
	)

	return (cpuAvailability + memoryAvailability) / 2
}
