package collector

import (
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

const listWatchersFixture = `watcher=192.0.2.1:0/9029529117 client.2912896056 cookie=59662829605819
watcher=192.0.2.2:0/7112525448 client.3788598496 cookie=32683403498153
watcher=192.0.2.3:0/373516797 client.9722834063 cookie=42908777223600
`

var listWatchersFixtureWatchers = []Watcher{
	{Address: "192.0.2.1:0/9029529117", Client: "client.2912896056", Cookie: "59662829605819"},
	{Address: "192.0.2.2:0/7112525448", Client: "client.3788598496", Cookie: "32683403498153"},
	{Address: "192.0.2.3:0/373516797", Client: "client.9722834063", Cookie: "42908777223600"},
}

type mockWatcherProvider struct {
	watchers []Watcher
	err      error
}

func (m *mockWatcherProvider) GetWatchers(pool, object string) ([]Watcher, error) {
	// Simulate getting the watchers
	return m.watchers, m.err
}

func Test_parseListWatchers(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []Watcher
		wantErr bool
	}{
		{
			name:  "three-watchers",
			input: listWatchersFixture,
			want:  listWatchersFixtureWatchers,
		},
		{
			name:  "empty-output",
			input: "",
			want:  []Watcher{},
		},
		{
			name:  "only-blank-lines",
			input: "\n\n  \n",
			want:  []Watcher{},
		},
		{
			name:  "trailing-and-interleaved-blank-lines",
			input: "\n" + listWatchersFixture + "\n\n",
			want:  listWatchersFixtureWatchers,
		},
		{
			name:  "no-trailing-newline",
			input: "watcher=192.0.2.1:0/9029529117 client.2912896056 cookie=59662829605819",
			want:  listWatchersFixtureWatchers[:1],
		},
		{
			name:    "malformed-line",
			input:   listWatchersFixture + "error: something went wrong\n",
			wantErr: true,
		},
		{
			name:    "missing-field",
			input:   "watcher=192.0.2.1:0/9029529117 cookie=59662829605819\n",
			wantErr: true,
		},
		{
			name:    "wrong-prefixes",
			input:   "192.0.2.1:0/9029529117 client.2912896056 59662829605819\n",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseListWatchers([]byte(tt.input))
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCollector_collectLeaderWatcherMetrics(t *testing.T) {
	type fields struct {
		watcherProvider WatcherProvider
	}
	type args struct {
		pool string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "three-watchers",
			fields: fields{
				watcherProvider: &mockWatcherProvider{watchers: listWatchersFixtureWatchers},
			},
			args: args{
				pool: "pool1",
			},
			want: 3,
		},
		{
			name: "zero-watchers",
			fields: fields{
				watcherProvider: &mockWatcherProvider{watchers: []Watcher{}},
			},
			args: args{
				pool: "pool1",
			},
			want: 0,
		},
		{
			name: "provider-error",
			fields: fields{
				watcherProvider: &mockWatcherProvider{err: errors.New("rados failed")},
			},
			args: args{
				pool: "pool1",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{
				pools:           []string{tt.args.pool},
				watcherProvider: tt.fields.watcherProvider,
			}
			got, err := c.collectLeaderWatcherMetrics(tt.args.pool)
			if (err != nil) != tt.wantErr {
				t.Errorf("Collector.collectLeaderWatcherMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.Nil(t, got)
				return
			}
			metrics := []prometheus.Metric{
				prometheus.MustNewConstMetric(
					prometheus.NewDesc("rbd_mirror_leader_watchers", "Count of watchers on the rbd_mirror_leader object", []string{"pool"}, nil),
					prometheus.GaugeValue,
					tt.want,
					tt.args.pool,
				),
			}
			assert.ElementsMatch(t, metrics, got)
		})
	}
}
