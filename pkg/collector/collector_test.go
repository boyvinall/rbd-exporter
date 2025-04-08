package collector

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

type mockPoolStatusProvider struct {
	PoolStatus
}

func (m *mockPoolStatusProvider) GetPoolStatus(pool string) (PoolStatus, error) {
	// Simulate getting the pool status
	return m.PoolStatus, nil
}

func TestCollector_collectPoolMetrics(t *testing.T) {
	type fields struct {
		pools    []string
		provider PoolStatusProvider
	}
	type args struct {
		pool string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]float64
		wantErr bool
	}{
		{
			name: "return-data-with-zero-values",
			fields: fields{
				pools: []string{"pool1"},
				provider: &mockPoolStatusProvider{
					PoolStatus: PoolStatus{
						Summary: PoolStatusSummary{
							Health:       "OK",
							DaemonHealth: "OK",
							ImageHealth:  "OK",
							States: map[string]int{
								"replaying": 7,
								"stopped":   6677,
							},
						},
					},
				},
			},
			args: args{
				pool: "pool1",
			},
			want: map[string]float64{
				"replaying":       7,
				"stopped":         6677,
				"starting_replay": 0,
				"stopping_replay": 0,
				"down+unknown":    0,
				"unknown":         0,
				"syncing":         0,
			},
		},
		{
			name: "return-all-data",
			fields: fields{
				pools: []string{"pool1"},
				provider: &mockPoolStatusProvider{
					PoolStatus: PoolStatus{
						Summary: PoolStatusSummary{
							Health:       "OK",
							DaemonHealth: "OK",
							ImageHealth:  "OK",
							States: map[string]int{
								"replaying":       123,
								"stopped":         456,
								"starting_replay": 789,
								"stopping_replay": 101112,
								"down+unknown":    131415,
								"unknown":         1,
								"syncing":         2,
							},
						},
					},
				},
			},
			args: args{
				pool: "pool1",
			},
			want: map[string]float64{
				"replaying":       123,
				"stopped":         456,
				"starting_replay": 789,
				"stopping_replay": 101112,
				"down+unknown":    131415,
				"unknown":         1,
				"syncing":         2,
			},
		},
		{
			name: "also-return-data-with-unknown-states",
			fields: fields{
				pools: []string{"pool1"},
				provider: &mockPoolStatusProvider{
					PoolStatus: PoolStatus{
						Summary: PoolStatusSummary{
							Health:       "OK",
							DaemonHealth: "OK",
							ImageHealth:  "OK",
							States: map[string]int{
								"replaying": 7,
								"foo":       6677,
							},
						},
					},
				},
			},
			args: args{
				pool: "pool1",
			},
			want: map[string]float64{
				"replaying":       7,
				"stopped":         0,
				"starting_replay": 0,
				"stopping_replay": 0,
				"down+unknown":    0,
				"unknown":         0,
				"syncing":         0,
				"foo":             6677,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{
				pools:    tt.fields.pools,
				provider: tt.fields.provider,
			}
			got, err := c.collectPoolMetrics(tt.args.pool)
			if (err != nil) != tt.wantErr {
				t.Errorf("Collector.collectPoolMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var metrics []prometheus.Metric
			for k, v := range tt.want {
				metrics = append(metrics, prometheus.MustNewConstMetric(
					prometheus.NewDesc("rbd_mirror_pool_status_state", "Count of RBD mirror pool states", []string{"pool", "state"}, nil),
					prometheus.GaugeValue,
					v,
					tt.args.pool,
					k,
				))
			}
			assert.ElementsMatch(t, metrics, got)
		})
	}
}
