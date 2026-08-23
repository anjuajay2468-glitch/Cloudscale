package scheduler

import (
	"fmt"
	"sync"
	"time"
)

type PlacementStatus string

const (
	PlacementPending PlacementStatus = "PENDING"
	PlacementRunning PlacementStatus = "RUNNING"
	PlacementFailed  PlacementStatus = "FAILED"
)

type Placement struct {
	WorkloadID string
	NodeID     string
	Status     PlacementStatus
	StartedAt  time.Time
	Error      string
}

type PlacementManager struct {
	mu         sync.Mutex
	placements map[string]Placement
}

func NewPlacementManager() *PlacementManager {
	return &PlacementManager{
		placements: make(map[string]Placement),
	}
}

func (p *PlacementManager) Place(
	workload Workload,
	decision SchedulingDecision,
) (Placement, error) {

	if workload.ID == "" {
		return Placement{}, fmt.Errorf(
			"workload ID is required",
		)
	}

	if decision.WorkloadID != workload.ID {
		return Placement{}, fmt.Errorf(
			"scheduling decision does not match workload",
		)
	}

	if decision.NodeID == "" {
		return Placement{}, fmt.Errorf(
			"scheduled node is required",
		)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.placements[workload.ID]; exists {
		return Placement{}, fmt.Errorf(
			"workload %s is already placed",
			workload.ID,
		)
	}

	placement := Placement{
		WorkloadID: workload.ID,
		NodeID:     decision.NodeID,
		Status:     PlacementRunning,
		StartedAt:  time.Now(),
	}

	p.placements[workload.ID] = placement

	return placement, nil
}

func (p *PlacementManager) Get(
	workloadID string,
) (Placement, bool) {

	p.mu.Lock()
	defer p.mu.Unlock()

	placement, ok := p.placements[workloadID]

	return placement, ok
}

func (p *PlacementManager) List() []Placement {
	p.mu.Lock()
	defer p.mu.Unlock()

	result := make([]Placement, 0, len(p.placements))

	for _, placement := range p.placements {
		result = append(result, placement)
	}

	return result
}
