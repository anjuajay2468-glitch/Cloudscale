package scheduler

import (
	"fmt"
	"sync"
	"time"
)

type SchedulingFeedback struct {
	WorkloadID string
	NodeID     string

	RequestedCPU    int64
	ObservedCPU     int64
	RequestedMemory int64
	ObservedMemory  int64

	CPUErrorPercent    float64
	MemoryErrorPercent float64

	RecordedAt time.Time
}

type FeedbackCollector struct {
	mu      sync.Mutex
	records map[string]SchedulingFeedback
}

func NewFeedbackCollector() *FeedbackCollector {
	return &FeedbackCollector{
		records: make(map[string]SchedulingFeedback),
	}
}

func (f *FeedbackCollector) Record(
	workload Workload,
	placement Placement,
	observed NodeResources,
) (SchedulingFeedback, error) {

	if workload.ID == "" {
		return SchedulingFeedback{}, fmt.Errorf(
			"workload ID is required",
		)
	}

	if placement.WorkloadID != workload.ID {
		return SchedulingFeedback{}, fmt.Errorf(
			"placement does not match workload",
		)
	}

	if placement.NodeID != observed.NodeID {
		return SchedulingFeedback{}, fmt.Errorf(
			"observed node does not match placement",
		)
	}

	feedback := SchedulingFeedback{
		WorkloadID:      workload.ID,
		NodeID:          placement.NodeID,
		RequestedCPU:    workload.CPURequestMillis,
		ObservedCPU:     observed.CPUUsedMillis,
		RequestedMemory: workload.MemoryRequestMB,
		ObservedMemory:  observed.MemoryUsedMB,
		RecordedAt:      time.Now(),
	}

	feedback.CPUErrorPercent = percentageError(
		feedback.RequestedCPU,
		feedback.ObservedCPU,
	)

	feedback.MemoryErrorPercent = percentageError(
		feedback.RequestedMemory,
		feedback.ObservedMemory,
	)

	f.mu.Lock()
	defer f.mu.Unlock()

	f.records[workload.ID] = feedback

	return feedback, nil
}

func (f *FeedbackCollector) Get(
	workloadID string,
) (SchedulingFeedback, bool) {

	f.mu.Lock()
	defer f.mu.Unlock()

	record, ok := f.records[workloadID]

	return record, ok
}

func (f *FeedbackCollector) List() []SchedulingFeedback {
	f.mu.Lock()
	defer f.mu.Unlock()

	result := make([]SchedulingFeedback, 0, len(f.records))

	for _, record := range f.records {
		result = append(result, record)
	}

	return result
}

func percentageError(requested, observed int64) float64 {
	if requested == 0 {
		if observed == 0 {
			return 0
		}

		return 100
	}

	difference := observed - requested

	if difference < 0 {
		difference = -difference
	}

	return float64(difference) / float64(requested) * 100
}
