package scheduler

import (
	"os"
	"testing"
)

func TestMetricsCollectorRequiresKubernetes(t *testing.T) {
	oldHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	oldPort := os.Getenv("KUBERNETES_SERVICE_PORT")

	defer os.Setenv("KUBERNETES_SERVICE_HOST", oldHost)
	defer os.Setenv("KUBERNETES_SERVICE_PORT", oldPort)

	os.Unsetenv("KUBERNETES_SERVICE_HOST")
	os.Unsetenv("KUBERNETES_SERVICE_PORT")

	collector := NewMetricsCollector("cloudscale")

	_, err := collector.CollectPodMetrics()

	if err == nil {
		t.Fatal("expected Kubernetes environment error")
	}
}
