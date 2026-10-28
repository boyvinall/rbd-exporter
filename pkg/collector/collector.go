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

// DaemonStatusProvider is an interface that defines a method to get the status of cephadm-managed daemons
// of a given type, as reported by `ceph orch ps`
type DaemonStatusProvider interface {
	GetDaemonStatuses(daemonType string) ([]DaemonStatus, error)
}

// WatcherProvider is an interface that defines a method to get the watchers of a RADOS object,
// as reported by `rados listwatchers`
type WatcherProvider interface {
	GetWatchers(pool, object string) ([]Watcher, error)
}

// Collector is a Prometheus exporter for RBD mirror pool status
type Collector struct {
	pools           []string
	daemonTypes     []string
	poolProvider    PoolStatusProvider
	daemonProvider  DaemonStatusProvider
	watcherProvider WatcherProvider
}

// New creates a new Collector instance
func New(pools, daemonTypes []string, poolProvider PoolStatusProvider, daemonProvider DaemonStatusProvider, watcherProvider WatcherProvider) *Collector {
	return &Collector{
		pools:           pools,
		daemonTypes:     daemonTypes,
		poolProvider:    poolProvider,
		daemonProvider:  daemonProvider,
		watcherProvider: watcherProvider,
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

// DaemonStatus represents a single daemon
// It is deserialised from the JSON output of the `ceph orch ps` command
type DaemonStatus struct {
	DaemonName string `json:"daemon_name"`
	Hostname   string `json:"hostname"`
	StatusDesc string `json:"status_desc"`
}

// Watcher represents a single watcher of a RADOS object
// It is parsed from the output of the `rados listwatchers` command
type Watcher struct {
	Address string
	Client  string
	Cookie  string
}

func (c *Collector) collectPoolMetrics(pool string) ([]prometheus.Metric, error) {
	s, err := c.poolProvider.GetPoolStatus(pool)
	if err != nil {
		return nil, err
	}

	// For each known-possible state, ensure it exists in the summary and initialize it to 0 if it doesn't
	// We could do this by hardcoding the individual states in the struct definition, but it's nice to be able
	// to know about new states without having to update the code.
	for _, state := range []string{"replaying", "starting_replay", "stopping_replay", "stopped", "down+unknown", "unknown", "syncing"} {
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

func (c *Collector) collectDaemonMetrics(daemonType string) ([]prometheus.Metric, error) {
	daemons, err := c.daemonProvider.GetDaemonStatuses(daemonType)
	if err != nil {
		return nil, err
	}

	// As with pool states, ensure each known-possible status is reported even when no daemon has it,
	// while still passing through any status we don't know about.
	counts := map[string]int{}
	for _, status := range []string{"running", "starting", "stopped", "error", "unknown"} {
		counts[status] = 0
	}
	for _, d := range daemons {
		counts[d.StatusDesc]++
	}

	metrics := make([]prometheus.Metric, 0, len(counts))
	for status, count := range counts {
		metrics = append(metrics, prometheus.MustNewConstMetric(
			prometheus.NewDesc("ceph_daemons", "Count of cephadm-managed Ceph daemons by type and status", []string{"daemon_type", "status"}, nil),
			prometheus.GaugeValue,
			float64(count),
			daemonType,
			status,
		))
	}

	return metrics, nil
}

func (c *Collector) collectLeaderWatcherMetrics(pool string) ([]prometheus.Metric, error) {
	watchers, err := c.watcherProvider.GetWatchers(pool, "rbd_mirror_leader")
	if err != nil {
		return nil, err
	}

	return []prometheus.Metric{
		prometheus.MustNewConstMetric(
			prometheus.NewDesc("rbd_mirror_leader_watchers", "Count of watchers on the rbd_mirror_leader object", []string{"pool"}, nil),
			prometheus.GaugeValue,
			float64(len(watchers)),
			pool,
		),
	}, nil
}

// Collect implements the [prometheus.Collector] interface
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	send := func(metrics []prometheus.Metric, err error) {
		if err != nil {
			ch <- prometheus.NewInvalidMetric(prometheus.NewDesc("rbd_exporter_error", "Error collecting metrics", nil, nil), err)
			return
		}

		for _, metric := range metrics {
			ch <- metric
		}
	}

	for _, pool := range c.pools {
		// Collect metrics for each pool
		send(c.collectPoolMetrics(pool))
		send(c.collectLeaderWatcherMetrics(pool))
	}

	for _, daemonType := range c.daemonTypes {
		send(c.collectDaemonMetrics(daemonType))
	}
}

// Describe implements the [prometheus.Collector] interface
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}
