package scheduler

import (
	"testing"
	"time"
)

func TestSchedulingFeedback(t *testing.T) {
	collector := NewFeedbackCollector()

	workload := Workload{
		ID:               "workload-1",
		CPURequestMillis: 500,
		MemoryRequestMB:  256,
		CreatedAt:        time.Now(),
	}

	placement := Placement{
		WorkloadID: workload.ID,
		NodeID:     "node2",
		Status:     PlacementRunning,
		StartedAt:  time.Now(),
	}

	observed := NodeResources{
		NodeID:         "node2",
		CPUUsedMillis:  600,
		MemoryUsedMB:   300,
		CPUTotalMillis: 2000,
		MemoryTotalMB:  4096,
		Healthy:        true,
	}

	feedback, err := collector.Record(
		workload,
		placement,
		observed,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if feedback.WorkloadID != "workload-1" {
		t.Fatalf(
			"expected workload-1, got %s",
			feedback.WorkloadID,
		)
	}

	if feedback.NodeID != "node2" {
		t.Fatalf(
			"expected node2, got %s",
			feedback.NodeID,
		)
	}

	if feedback.CPUErrorPercent != 20 {
		t.Fatalf(
			"expected CPU error 20%%, got %.2f%%",
			feedback.CPUErrorPercent,
		)
	}

	if feedback.MemoryErrorPercent != 17.1875 {
		t.Fatalf(
			"expected memory error 17.1875%%, got %.4f%%",
			feedback.MemoryErrorPercent,
		)
	}
}

func TestFeedbackRejectsWrongNode(t *testing.T) {
	collector := NewFeedbackCollector()

	workload := Workload{
		ID: "workload-1",
	}

	placement := Placement{
		WorkloadID: workload.ID,
		NodeID:     "node1",
	}

	observed := NodeResources{
		NodeID: "node2",
	}

	_, err := collector.Record(
		workload,
		placement,
		observed,
	)

	if err == nil {
		t.Fatal("expected node mismatch error")
	}
}

func TestPercentageError(t *testing.T) {
	tests := []struct {
		name     string
		request  int64
		observed int64
		expected float64
	}{
		{
			name:     "equal",
			request:  100,
			observed: 100,
			expected: 0,
		},
		{
			name:     "higher",
			request:  100,
			observed: 150,
			expected: 50,
		},
		{
			name:     "lower",
			request:  100,
			observed: 50,
			expected: 50,
		},
		{
			name:     "zero request and zero usage",
			request:  0,
			observed: 0,
			expected: 0,
		},
		{
			name:     "zero request with usage",
			request:  0,
			observed: 100,
			expected: 100,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := percentageError(
				test.request,
				test.observed,
			)

			if got != test.expected {
				t.Fatalf(
					"expected %.2f, got %.2f",
					test.expected,
					got,
				)
			}
		})
	}
}
