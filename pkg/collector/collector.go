// Package collector provides a Prometheus collector for RBD mirror pool status
package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

// PoolStatusProvider is an interface that defines a method to get the status of an RBD mirror pool
// This interface is used to decouple the collector from the actual implementation of getting the pool status
// which allows for easier testing and mocking of the pool status retrieval
type PoolStatusProvider interface {
	GetPoolStatus(pool string) (PoolStatus, error)
}

// Collector is a Prometheus exporter for RBD mirror pool status
type Collector struct {
	pools    []string
	provider PoolStatusProvider
}

// New creates a new Collector instance
func New(pools []string, provider PoolStatusProvider) *Collector {
	return &Collector{
		pools:    pools,
		provider: provider,
	}
}

// PoolStatusSummary is used by [PoolStatus]
// Having a separate struct for the summary allows us to more easily initialize local variables
type PoolStatusSummary struct {
	Health       string         `json:"health"`
	DaemonHealth string         `json:"daemon_health,omitempty"`
	ImageHealth  string         `json:"image_health,omitempty"`
	States       map[string]int `json:"states"`
}

// PoolStatus represents the status of an RBD mirror pool
// It is deserialises the JSON output from the `rbd mirror pool status` command
type PoolStatus struct {
	Summary PoolStatusSummary `json:"summary"`
}

func (c *Collector) collectPoolMetrics(pool string) ([]prometheus.Metric, error) {
	s, err := c.provider.GetPoolStatus(pool)
	if err != nil {
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
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	for _, pool := range c.pools {
		// Collect metrics for each pool
		metrics, err := c.collectPoolMetrics(pool)
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
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}
