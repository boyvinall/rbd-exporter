package main

import (
	"encoding/json"
	"log/slog"
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector is a Prometheus exporter for RBD mirror pool status
type Collector struct {
	Pools []string
}

type poolStatusSummary struct {
	Health       string         `json:"health"`
	DaemonHealth string         `json:"daemon_health,omitempty"`
	ImageHealth  string         `json:"image_health,omitempty"`
	States       map[string]int `json:"states"`
}

type poolStatus struct {
	Summary poolStatusSummary `json:"summary"`
}

func collectMetrics(pool string) ([]prometheus.Metric, error) {
	// Simulate collecting metrics from the pool
	// In a real implementation, this would involve querying the RBD cluster
	// and collecting relevant metrics.

	// s := poolStatus{
	// 	Summary: poolStatusSummary{
	// 		Health: "OK",
	// 		States: map[string]int{
	// 			"replaying": 7,
	// 			"stopped":   6677,
	// 		},
	// 	},
	// }

	b, err := exec.Command("rbd", "mirror", "pool", "status", "--format", "json", pool).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			slog.Error("rbd mirror pool status", "pool", pool, "error", string(exitErr.Stderr))
		} else {
			slog.Error("rbd mirror pool status", "pool", pool, "error", err)
		}
		return nil, err
	}

	// Unmarshal the JSON output into the poolStatus struct
	var s poolStatus
	err = json.Unmarshal(b, &s)
	if err != nil {
		slog.Error("failed to unmarshal JSON", "error", err)
		return nil, err
	}

	// For each known-possible state, ensure it exists in the summary and initialize it to 0 if it doesn't
	// We could do this by hardcoding the individual states in the struct definition, but it's nice to be able
	// to know about new states without having to update the code.
	for _, state := range []string{"replaying", "starting_replay", "stopping_replay", "stopped", "down+unknown"} {
		s.Summary.States[state] += 0
	}

	metrics := make([]prometheus.Metric, 0, len(s.Summary.States))
	for state, count := range s.Summary.States {
		metrics = append(metrics, prometheus.MustNewConstMetric(
			prometheus.NewDesc("rbd_mirror_pool_status_state", "Count of RBD mirror pool states", []string{"pool", "state"}, nil),
			prometheus.GaugeValue,
			float64(count),
			pool,
			state,
		))
	}

	return metrics, nil
}

// Collect implements the [prometheus.Collector] interface
func (e *Collector) Collect(ch chan<- prometheus.Metric) {
	for _, pool := range e.Pools {
		// Collect metrics for each pool
		metrics, err := collectMetrics(pool)
		if err != nil {
			ch <- prometheus.NewInvalidMetric(prometheus.NewDesc("rbd_exporter_error", "Error collecting metrics", nil, nil), err)
			continue
		}

		for _, metric := range metrics {
			ch <- metric
		}
	}
}

// Describe implements the [prometheus.Collector] interface
func (e *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(e, ch)
}
