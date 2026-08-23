package scheduler

import (
	"testing"
)

func TestLearningScheduler(t *testing.T) {
	scheduler := NewLearningScheduler()

	cpu, memory := scheduler.Weights()

	if cpu != 0.5 || memory != 0.5 {
		t.Fatalf(
			"expected initial weights 0.5/0.5, got %.2f/%.2f",
			cpu,
			memory,
		)
	}

	feedback := SchedulingFeedback{
		WorkloadID:         "workload-1",
		NodeID:             "node1",
		CPUErrorPercent:    80,
		MemoryErrorPercent: 20,
	}

	scheduler.Learn(feedback)

	cpu, memory = scheduler.Weights()

	if cpu <= 0.5 {
		t.Fatalf(
			"expected CPU weight to increase, got %.3f",
			cpu,
		)
	}

	if memory >= 0.5 {
		t.Fatalf(
			"expected memory weight to decrease, got %.3f",
			memory,
		)
	}

	if cpu+memory < 0.999 || cpu+memory > 1.001 {
		t.Fatalf(
			"weights are not normalized: %.3f + %.3f",
			cpu,
			memory,
		)
	}
}

func TestLearningSchedulerSelectsHealthyNode(t *testing.T) {
	scheduler := NewLearningScheduler()

	nodes := []NodeResources{
		{
			NodeID:         "node1",
			CPUUsedMillis:  900,
			MemoryUsedMB:   400,
			CPUTotalMillis: 1000,
			MemoryTotalMB:  512,
			Healthy:        true,
		},
		{
			NodeID:         "node2",
			CPUUsedMillis:  100,
			MemoryUsedMB:   100,
			CPUTotalMillis: 1000,
			MemoryTotalMB:  512,
			Healthy:        true,
		},
		{
			NodeID:         "node3",
			CPUUsedMillis:  200,
			MemoryUsedMB:   150,
			CPUTotalMillis: 1000,
			MemoryTotalMB:  512,
			Healthy:        false,
		},
	}

	node, score, err := scheduler.SelectNode(
		nodes,
		nil,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NodeID != "node2" {
		t.Fatalf(
			"expected node2, got %s",
			node.NodeID,
		)
	}

	if score <= 0 {
		t.Fatalf(
			"expected positive score, got %.3f",
			score,
		)
	}
}
