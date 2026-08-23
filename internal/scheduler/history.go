package scheduler

import (
	"sync"
	"time"
)

type WorkloadHistory struct {
	WorkloadID string
	NodeID     string

	RequestedCPU    int64
	RequestedMemory int64

	NodeCPUUsed     int64
	NodeMemoryUsed  int64
	NodeCPUTotal    int64
	NodeMemoryTotal int64

	CPUErrorPercent    float64
	MemoryErrorPercent float64

	Score float64

	Timestamp time.Time
}

type HistoryStore struct {
	mu      sync.Mutex
	records []WorkloadHistory
}

func NewHistoryStore() *HistoryStore {
	return &HistoryStore{
		records: make([]WorkloadHistory, 0),
	}
}

func (h *HistoryStore) Add(
	workload Workload,
	decision SchedulingDecision,
	observed NodeResources,
	feedback SchedulingFeedback,
) WorkloadHistory {

	record := WorkloadHistory{
		WorkloadID: workload.ID,
		NodeID:     decision.NodeID,

		RequestedCPU:    workload.CPURequestMillis,
		RequestedMemory: workload.MemoryRequestMB,

		NodeCPUUsed:     observed.CPUUsedMillis,
		NodeMemoryUsed:  observed.MemoryUsedMB,
		NodeCPUTotal:    observed.CPUTotalMillis,
		NodeMemoryTotal: observed.MemoryTotalMB,

		CPUErrorPercent:    feedback.CPUErrorPercent,
		MemoryErrorPercent: feedback.MemoryErrorPercent,

		Score: decision.Score,

		Timestamp: time.Now(),
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.records = append(h.records, record)

	return record
}

func (h *HistoryStore) List() []WorkloadHistory {
	h.mu.Lock()
	defer h.mu.Unlock()

	result := make([]WorkloadHistory, len(h.records))
	copy(result, h.records)

	return result
}

func (h *HistoryStore) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	return len(h.records)
}
