package scheduler

import (
	"testing"
	"time"
)

func TestSelectNode(t *testing.T) {
	s := NewScheduler()

	workload := Workload{
		ID:               "test-workload",
		CPURequestMillis: 500,
		MemoryRequestMB:  1024,
		Priority:         1,
		CreatedAt:        time.Now(),
	}

	nodes := []NodeResources{
		{
			NodeID:         "node1",
			CPUTotalMillis: 4000,
			CPUUsedMillis:  3000,
			MemoryTotalMB:  8192,
			MemoryUsedMB:   4000,
			LatencyMs:      10,
			Healthy:        true,
		},
		{
			NodeID:         "node2",
			CPUTotalMillis: 4000,
			CPUUsedMillis:  1000,
			MemoryTotalMB:  8192,
			MemoryUsedMB:   3000,
			LatencyMs:      20,
			Healthy:        true,
		},
		{
			NodeID:         "node3",
			CPUTotalMillis: 4000,
			CPUUsedMillis:  3500,
			MemoryTotalMB:  8192,
			MemoryUsedMB:   5000,
			LatencyMs:      5,
			Healthy:        true,
		},
	}

	selected, err := s.SelectNode(workload, nodes)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if selected.NodeID != "node2" {
		t.Fatalf(
			"expected node2, got %s",
			selected.NodeID,
		)
	}
}

func TestUnhealthyNodeRejected(t *testing.T) {
	s := NewScheduler()

	workload := Workload{
		ID:               "test-workload",
		CPURequestMillis: 500,
		MemoryRequestMB:  1024,
	}

	nodes := []NodeResources{
		{
			NodeID:         "node1",
			CPUTotalMillis: 4000,
			CPUUsedMillis:  500,
			MemoryTotalMB:  8192,
			MemoryUsedMB:   1000,
			Healthy:        false,
		},
		{
			NodeID:         "node2",
			CPUTotalMillis: 4000,
			CPUUsedMillis:  3000,
			MemoryTotalMB:  8192,
			MemoryUsedMB:   7000,
			Healthy:        true,
		},
	}

	selected, err := s.SelectNode(workload, nodes)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if selected.NodeID != "node2" {
		t.Fatalf(
			"expected node2, got %s",
			selected.NodeID,
		)
	}
}
