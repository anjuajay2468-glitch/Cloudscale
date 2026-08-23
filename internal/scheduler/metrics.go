package scheduler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type PodMetric struct {
	Name     string
	CPUMilli int64
	MemoryMB int64
}

type MetricsCollector struct {
	Client    *http.Client
	Namespace string
}

func NewMetricsCollector(namespace string) *MetricsCollector {
	return &MetricsCollector{
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
		Namespace: namespace,
	}
}

type metricsResponse struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Containers []struct {
			Name  string `json:"name"`
			Usage struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
			} `json:"usage"`
		} `json:"containers"`
	} `json:"items"`
}

func (c *MetricsCollector) CollectPodMetrics() ([]PodMetric, error) {
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")

	if host == "" || port == "" {
		return nil, fmt.Errorf(
			"Kubernetes service environment variables are unavailable",
		)
	}

	url := fmt.Sprintf(
		"https://%s:%s/apis/metrics.k8s.io/v1beta1/namespaces/%s/pods",
		host,
		port,
		c.Namespace,
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	token, err := os.ReadFile(
		"/var/run/secrets/kubernetes.io/serviceaccount/token",
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read Kubernetes service account token: %w",
			err,
		)
	}

	req.Header.Set(
		"Authorization",
		"Bearer "+strings.TrimSpace(string(token)),
	)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query Metrics API: %w",
			err,
		)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"Metrics API returned HTTP %d",
			resp.StatusCode,
		)
	}

	var payload metricsResponse

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf(
			"failed to decode Metrics API response: %w",
			err,
		)
	}

	result := make([]PodMetric, 0, len(payload.Items))

	for _, item := range payload.Items {
		var cpuMilli int64
		var memoryBytes int64

		for _, container := range item.Containers {
			cpu, err := parseCPU(container.Usage.CPU)
			if err != nil {
				return nil, err
			}

			memory, err := parseMemory(container.Usage.Memory)
			if err != nil {
				return nil, err
			}

			cpuMilli += cpu
			memoryBytes += memory
		}

		result = append(result, PodMetric{
			Name:     item.Metadata.Name,
			CPUMilli: cpuMilli,
			MemoryMB: memoryBytes / (1024 * 1024),
		})
	}

	return result, nil
}

func parseCPU(value string) (int64, error) {
	if strings.HasSuffix(value, "n") {
		nano, err := strconv.ParseInt(
			strings.TrimSuffix(value, "n"),
			10,
			64,
		)
		if err != nil {
			return 0, err
		}

		return nano / 1_000_000, nil
	}

	if strings.HasSuffix(value, "u") {
		micro, err := strconv.ParseInt(
			strings.TrimSuffix(value, "u"),
			10,
			64,
		)
		if err != nil {
			return 0, err
		}

		return micro / 1_000, nil
	}

	if strings.HasSuffix(value, "m") {
		return strconv.ParseInt(
			strings.TrimSuffix(value, "m"),
			10,
			64,
		)
	}

	return 0, fmt.Errorf("unsupported CPU quantity: %s", value)
}

func parseMemory(value string) (int64, error) {
	multipliers := map[string]int64{
		"Ki": 1024,
		"Mi": 1024 * 1024,
		"Gi": 1024 * 1024 * 1024,
	}

	for suffix, multiplier := range multipliers {
		if strings.HasSuffix(value, suffix) {
			number, err := strconv.ParseFloat(
				strings.TrimSuffix(value, suffix),
				64,
			)
			if err != nil {
				return 0, err
			}

			return int64(number * float64(multiplier)), nil
		}
	}

	return strconv.ParseInt(value, 10, 64)
}
func (c *MetricsCollector) CollectNodeResources() ([]NodeResources, error) {
	metrics, err := c.CollectPodMetrics()
	if err != nil {
		return nil, err
	}

	nodes := make([]NodeResources, 0)

	for _, metric := range metrics {
		if !strings.HasPrefix(metric.Name, "node") {
			continue
		}

		nodes = append(nodes, NodeResources{
			NodeID:         strings.Split(metric.Name, "-")[0],
			CPUUsedMillis:  metric.CPUMilli,
			MemoryUsedMB:   metric.MemoryMB,
			CPUTotalMillis: 1000,
			MemoryTotalMB:  512,
			Healthy:        true,
		})
	}

	return nodes, nil
}
