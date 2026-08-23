package scheduler

import "time"

// Workload represents a unit of work that CloudScale
// needs to place on one of its nodes.
type Workload struct {
	ID string

	// Estimated resource requirements.
	CPURequestMillis int64
	MemoryRequestMB  int64

	// Scheduling characteristics.
	Priority int

	// Higher values mean the workload is more sensitive
	// to network latency.
	LatencySensitivity float64

	// Maximum acceptable latency in milliseconds.
	MaxLatencyMs float64

	// Creation time is useful for measuring scheduling delay.
	CreatedAt time.Time
}

// NodeResources describes the currently available
// resources on a CloudScale node.
type NodeResources struct {
	NodeID string

	CPUTotalMillis int64
	CPUUsedMillis  int64

	MemoryTotalMB int64
	MemoryUsedMB  int64

	// Network latency measured from the scheduler
	// to this node.
	LatencyMs float64

	// Whether the node is currently considered healthy.
	Healthy bool
}

// AvailableCPU returns currently available CPU.
func (n NodeResources) AvailableCPU() int64 {
	available := n.CPUTotalMillis - n.CPUUsedMillis

	if available < 0 {
		return 0
	}

	return available
}

// AvailableMemory returns currently available memory.
func (n NodeResources) AvailableMemory() int64 {
	available := n.MemoryTotalMB - n.MemoryUsedMB

	if available < 0 {
		return 0
	}

	return available
}

// CanRun determines whether a node has enough resources
// to run the workload.
func (n NodeResources) CanRun(w Workload) bool {
	if !n.Healthy {
		return false
	}

	if n.AvailableCPU() < w.CPURequestMillis {
		return false
	}

	if n.AvailableMemory() < w.MemoryRequestMB {
		return false
	}

	if w.MaxLatencyMs > 0 &&
		n.LatencyMs > w.MaxLatencyMs {
		return false
	}

	return true
}
