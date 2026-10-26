package collector

import (
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

// orchPSRBDMirrorFixture is real (anonymized) output of `ceph orch ps --daemon-type rbd-mirror --format json`
// from a Ceph 18.2.7 cluster running three rbd-mirror daemons.
const orchPSRBDMirrorFixture = `[
	{
		"container_id": "7780ecc176f2",
		"container_image_digests": [
			"quay.io/ceph/ceph@sha256:1b9158ce28975f95def6a0ad459fa19f1336506074267a4b47c1bd914a00fec0",
			"quay.io/ceph/ceph@sha256:7d8bb82696d5d9cbeae2a2828dc12b6835aa2dded890fa3ac5a733cb66b72b1c"
		],
		"container_image_id": "0f5473a1e726b0feaff0f41f8de8341c0a94f60365d4584f4c10bd6b40d44bc1",
		"container_image_name": "quay.io/ceph/ceph@sha256:1b9158ce28975f95def6a0ad459fa19f1336506074267a4b47c1bd914a00fec0",
		"cpu_percentage": "51.71%",
		"created": "2025-05-02T08:19:00.752595Z",
		"daemon_id": "server1.vbstif",
		"daemon_name": "rbd-mirror.server1.vbstif",
		"daemon_type": "rbd-mirror",
		"events": [
			"2026-10-23T17:23:05.427383Z daemon:rbd-mirror.server1.vbstif [INFO] \"restart rbd-mirror.server1.vbstif from host 'server1'\"",
			"2026-10-23T17:59:28.961283Z daemon:rbd-mirror.server1.vbstif [INFO] \"stop rbd-mirror.server1.vbstif from host 'server1'\"",
			"2026-10-23T18:01:40.770792Z daemon:rbd-mirror.server1.vbstif [INFO] \"start rbd-mirror.server1.vbstif from host 'server1'\""
		],
		"hostname": "server1",
		"is_active": false,
		"last_refresh": "2026-10-26T09:42:58.562737Z",
		"memory_usage": 2765958938,
		"ports": [],
		"service_name": "rbd-mirror",
		"started": "2026-10-23T18:01:40.667434Z",
		"status": 1,
		"status_desc": "running",
		"systemd_unit": "ceph-4223779f-06e3-4bb3-ac66-9c556804a1bc@rbd-mirror.server1.vbstif",
		"version": "18.2.7"
	},
	{
		"container_id": "e8da50da202c",
		"container_image_digests": [
			"quay.io/ceph/ceph@sha256:7d8bb82696d5d9cbeae2a2828dc12b6835aa2dded890fa3ac5a733cb66b72b1c",
			"quay.io/ceph/ceph@sha256:1b9158ce28975f95def6a0ad459fa19f1336506074267a4b47c1bd914a00fec0"
		],
		"container_image_id": "0f5473a1e726b0feaff0f41f8de8341c0a94f60365d4584f4c10bd6b40d44bc1",
		"container_image_name": "quay.io/ceph/ceph@sha256:1b9158ce28975f95def6a0ad459fa19f1336506074267a4b47c1bd914a00fec0",
		"cpu_percentage": "23.65%",
		"created": "2026-09-23T07:00:48.831694Z",
		"daemon_id": "server2.qfvsrl",
		"daemon_name": "rbd-mirror.server2.qfvsrl",
		"daemon_type": "rbd-mirror",
		"events": [
			"2026-10-26T07:36:51.648679Z daemon:rbd-mirror.server2.qfvsrl [INFO] \"restart rbd-mirror.server2.qfvsrl from host 'server2'\"",
			"2026-10-23T17:59:40.046978Z daemon:rbd-mirror.server2.qfvsrl [INFO] \"stop rbd-mirror.server2.qfvsrl from host 'server2'\"",
			"2026-10-23T18:01:41.969804Z daemon:rbd-mirror.server2.qfvsrl [INFO] \"start rbd-mirror.server2.qfvsrl from host 'server2'\""
		],
		"hostname": "server2",
		"is_active": false,
		"last_refresh": "2026-10-26T09:49:16.508208Z",
		"memory_usage": 2064805527,
		"ports": [],
		"service_name": "rbd-mirror",
		"started": "2026-10-26T07:36:51.542800Z",
		"status": 1,
		"status_desc": "running",
		"systemd_unit": "ceph-4223779f-06e3-4bb3-ac66-9c556804a1bc@rbd-mirror.server2.qfvsrl",
		"version": "18.2.7"
	},
	{
		"container_id": "cfb1c4dfc58d",
		"container_image_digests": [
			"quay.io/ceph/ceph@sha256:7d8bb82696d5d9cbeae2a2828dc12b6835aa2dded890fa3ac5a733cb66b72b1c",
			"quay.io/ceph/ceph@sha256:1b9158ce28975f95def6a0ad459fa19f1336506074267a4b47c1bd914a00fec0"
		],
		"container_image_id": "0f5473a1e726b0feaff0f41f8de8341c0a94f60365d4584f4c10bd6b40d44bc1",
		"container_image_name": "quay.io/ceph/ceph@sha256:1b9158ce28975f95def6a0ad459fa19f1336506074267a4b47c1bd914a00fec0",
		"cpu_percentage": "31.33%",
		"created": "2026-09-23T07:00:46.860871Z",
		"daemon_id": "server3.qpnuvs",
		"daemon_name": "rbd-mirror.server3.qpnuvs",
		"daemon_type": "rbd-mirror",
		"events": [
			"2026-10-26T07:37:03.422641Z daemon:rbd-mirror.server3.qpnuvs [INFO] \"restart rbd-mirror.server3.qpnuvs from host 'server3'\"",
			"2026-10-23T17:59:43.887097Z daemon:rbd-mirror.server3.qpnuvs [INFO] \"stop rbd-mirror.server3.qpnuvs from host 'server3'\"",
			"2026-10-23T18:01:42.985366Z daemon:rbd-mirror.server3.qpnuvs [INFO] \"start rbd-mirror.server3.qpnuvs from host 'server3'\""
		],
		"hostname": "server3",
		"is_active": false,
		"last_refresh": "2026-10-26T09:49:17.246095Z",
		"memory_usage": 1915555414,
		"ports": [],
		"service_name": "rbd-mirror",
		"started": "2026-10-26T07:37:03.328863Z",
		"status": 1,
		"status_desc": "running",
		"systemd_unit": "ceph-4223779f-06e3-4bb3-ac66-9c556804a1bc@rbd-mirror.server3.qpnuvs",
		"version": "18.2.7"
	}
]`

func Test_parseOrchPS(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []DaemonStatus
		wantErr bool
	}{
		{
			name:  "ceph-18.2.7-rbd-mirror-all-running",
			input: orchPSRBDMirrorFixture,
			want: []DaemonStatus{
				{DaemonName: "rbd-mirror.server1.vbstif", Hostname: "server1", StatusDesc: "running"},
				{DaemonName: "rbd-mirror.server2.qfvsrl", Hostname: "server2", StatusDesc: "running"},
				{DaemonName: "rbd-mirror.server3.qpnuvs", Hostname: "server3", StatusDesc: "running"},
			},
		},
		{
			name:  "empty-status-desc-becomes-unknown",
			input: `[{"daemon_name": "rbd-mirror.server1.vbstif", "hostname": "server1", "status": -2}]`,
			want: []DaemonStatus{
				{DaemonName: "rbd-mirror.server1.vbstif", Hostname: "server1", StatusDesc: "unknown"},
			},
		},
		{
			name:  "empty-array",
			input: `[]`,
			want:  []DaemonStatus{},
		},
		{
			name:    "invalid-json",
			input:   `[{"daemon_name": `,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOrchPS([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseOrchPS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

type mockDaemonStatusProvider struct {
	daemons []DaemonStatus
	err     error
}

func (m *mockDaemonStatusProvider) GetDaemonStatuses(daemonType string) ([]DaemonStatus, error) {
	// Simulate getting the daemon statuses
	return m.daemons, m.err
}

func TestCollector_collectDaemonMetrics(t *testing.T) {
	type fields struct {
		daemonProvider DaemonStatusProvider
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]float64
		wantErr bool
	}{
		{
			name: "all-running",
			fields: fields{
				daemonProvider: &mockDaemonStatusProvider{
					daemons: []DaemonStatus{
						{DaemonName: "rbd-mirror.server1.vbstif", Hostname: "server1", StatusDesc: "running"},
						{DaemonName: "rbd-mirror.server2.qfvsrl", Hostname: "server2", StatusDesc: "running"},
						{DaemonName: "rbd-mirror.server3.qpnuvs", Hostname: "server3", StatusDesc: "running"},
					},
				},
			},
			want: map[string]float64{
				"running":  3,
				"starting": 0,
				"stopped":  0,
				"error":    0,
				"unknown":  0,
			},
		},
		{
			name: "mixed-statuses",
			fields: fields{
				daemonProvider: &mockDaemonStatusProvider{
					daemons: []DaemonStatus{
						{DaemonName: "rbd-mirror.server1.vbstif", Hostname: "server1", StatusDesc: "running"},
						{DaemonName: "rbd-mirror.server2.qfvsrl", Hostname: "server2", StatusDesc: "error"},
						{DaemonName: "rbd-mirror.server3.qpnuvs", Hostname: "server3", StatusDesc: "stopped"},
						{DaemonName: "rbd-mirror.server4.hhzmvv", Hostname: "server4", StatusDesc: "starting"},
						{DaemonName: "rbd-mirror.server5.rixyst", Hostname: "server5", StatusDesc: "unknown"},
						{DaemonName: "rbd-mirror.server6.ouxnxl", Hostname: "server6", StatusDesc: "running"},
					},
				},
			},
			want: map[string]float64{
				"running":  2,
				"starting": 1,
				"stopped":  1,
				"error":    1,
				"unknown":  1,
			},
		},
		{
			name: "also-return-data-with-unknown-statuses",
			fields: fields{
				daemonProvider: &mockDaemonStatusProvider{
					daemons: []DaemonStatus{
						{DaemonName: "rbd-mirror.server1.vbstif", Hostname: "server1", StatusDesc: "running"},
						{DaemonName: "rbd-mirror.server2.qfvsrl", Hostname: "server2", StatusDesc: "foo"},
					},
				},
			},
			want: map[string]float64{
				"running":  1,
				"starting": 0,
				"stopped":  0,
				"error":    0,
				"unknown":  0,
				"foo":      1,
			},
		},
		{
			name: "provider-error",
			fields: fields{
				daemonProvider: &mockDaemonStatusProvider{
					err: errors.New("ceph orch ps failed"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{
				daemonProvider: tt.fields.daemonProvider,
			}
			got, err := c.collectDaemonMetrics()
			if (err != nil) != tt.wantErr {
				t.Errorf("Collector.collectDaemonMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var metrics []prometheus.Metric
			for k, v := range tt.want {
				metrics = append(metrics, prometheus.MustNewConstMetric(
					prometheus.NewDesc("rbd_mirror_daemons", "Count of cephadm-managed rbd-mirror daemons by status", []string{"status"}, nil),
					prometheus.GaugeValue,
					v,
					k,
				))
			}
			assert.ElementsMatch(t, metrics, got)
		})
	}
}
