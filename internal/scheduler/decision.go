package scheduler

import (
	"fmt"
	"time"
)

type SchedulingDecision struct {
	WorkloadID string
	NodeID     string
	Score      float64
	CreatedAt  time.Time
	DecidedAt  time.Time
}

func (s *Scheduler) Schedule(
	workload Workload,
	nodes []NodeResources,
) (SchedulingDecision, error) {
	node, err := s.SelectNode(workload, nodes)
	if err != nil {
		return SchedulingDecision{}, err
	}

	now := time.Now()

	return SchedulingDecision{
		WorkloadID: workload.ID,
		NodeID:     node.NodeID,
		Score:      float64(node.AvailableCPU()),
		CreatedAt:  workload.CreatedAt,
		DecidedAt:  now,
	}, nil
}

func (s *Scheduler) ScheduleWithMetrics(
	workload Workload,
	collector *MetricsCollector,
) (SchedulingDecision, error) {

	nodes, err := collector.CollectNodeResources()
	if err != nil {
		return SchedulingDecision{}, fmt.Errorf(
			"failed to collect node resources: %w",
			err,
		)
	}

	return s.Schedule(workload, nodes)
}
