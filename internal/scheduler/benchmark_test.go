package scheduler

import (
	"testing"
	"time"
)

func TestCompareSchedulers(t *testing.T) {
	workload := Workload{
		ID:               "benchmark-1",
		CPURequestMillis: 500,
		MemoryRequestMB:  256,
		CreatedAt:        time.Now(),
	}

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
			CPUUsedMillis:  200,
			MemoryUsedMB:   100,
			CPUTotalMillis: 1000,
			MemoryTotalMB:  512,
			Healthy:        true,
		},
		{
			NodeID:         "node3",
			CPUUsedMillis:  500,
			MemoryUsedMB:   200,
			CPUTotalMillis: 1000,
			MemoryTotalMB:  512,
			Healthy:        true,
		},
	}

	scorer := NewIntelligentScorer()

	result, err := CompareSchedulers(
		workload,
		nodes,
		nil,
		scorer,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.BaselineNode != "node1" {
		t.Fatalf(
			"expected baseline to select node1, got %s",
			result.BaselineNode,
		)
	}

	if result.IntelligentNode != "node2" {
		t.Fatalf(
			"expected intelligent scheduler to select node2, got %s",
			result.IntelligentNode,
		)
	}

	if result.IntelligentScore <= result.BaselineScore {
		t.Fatalf(
			"expected intelligent score %.3f > baseline %.3f",
			result.IntelligentScore,
			result.BaselineScore,
		)
	}
}

func TestCompareSchedulersRejectsEmptyNodes(t *testing.T) {
	workload := Workload{
		ID: "benchmark-empty",
	}

	_, err := CompareSchedulers(
		workload,
		nil,
		nil,
		NewIntelligentScorer(),
	)

	if err == nil {
		t.Fatal("expected empty node error")
	}
}
