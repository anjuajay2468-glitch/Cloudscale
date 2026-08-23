package scheduler

import (
	"testing"
	"time"
)

func TestPlacement(t *testing.T) {
	manager := NewPlacementManager()

	workload := Workload{
		ID:               "workload-1",
		CPURequestMillis: 500,
		MemoryRequestMB:  256,
		CreatedAt:        time.Now(),
	}

	decision := SchedulingDecision{
		WorkloadID: workload.ID,
		NodeID:     "node2",
		Score:      3500,
		CreatedAt:  workload.CreatedAt,
		DecidedAt:  time.Now(),
	}

	placement, err := manager.Place(
		workload,
		decision,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if placement.WorkloadID != "workload-1" {
		t.Fatalf(
			"expected workload-1, got %s",
			placement.WorkloadID,
		)
	}

	if placement.NodeID != "node2" {
		t.Fatalf(
			"expected node2, got %s",
			placement.NodeID,
		)
	}

	if placement.Status != PlacementRunning {
		t.Fatalf(
			"expected RUNNING, got %s",
			placement.Status,
		)
	}

	loaded, ok := manager.Get("workload-1")

	if !ok {
		t.Fatal("expected placement to exist")
	}

	if loaded.NodeID != "node2" {
		t.Fatalf(
			"expected stored placement on node2, got %s",
			loaded.NodeID,
		)
	}
}

func TestDuplicatePlacementRejected(t *testing.T) {
	manager := NewPlacementManager()

	workload := Workload{
		ID: "workload-duplicate",
	}

	decision := SchedulingDecision{
		WorkloadID: workload.ID,
		NodeID:     "node1",
	}

	_, err := manager.Place(workload, decision)

	if err != nil {
		t.Fatalf("unexpected first placement error: %v", err)
	}

	_, err = manager.Place(workload, decision)

	if err == nil {
		t.Fatal("expected duplicate placement to fail")
	}
}

func TestMismatchedDecisionRejected(t *testing.T) {
	manager := NewPlacementManager()

	workload := Workload{
		ID: "workload-1",
	}

	decision := SchedulingDecision{
		WorkloadID: "workload-2",
		NodeID:     "node1",
	}

	_, err := manager.Place(workload, decision)

	if err == nil {
		t.Fatal("expected mismatched decision to fail")
	}
}
