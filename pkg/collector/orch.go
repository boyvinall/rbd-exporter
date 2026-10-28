package collector

import (
	"encoding/json"
	"log/slog"
	"os/exec"
)

// CephOrchPS implements the [DaemonStatusProvider] interface to exec the `ceph orch ps` command
type CephOrchPS struct{}

// GetDaemonStatuses implements the [DaemonStatusProvider] interface
//
// As with [RBDMirrorPoolStatus], this could be implemented against the mgr API,
// but for now we'll just exec out.
func (c *CephOrchPS) GetDaemonStatuses(daemonType string) ([]DaemonStatus, error) {
	// Execute the command and capture the output
	// For example: ceph orch ps --daemon-type rbd-mirror --format json

	b, err := exec.Command("ceph", "orch", "ps", "--daemon-type", daemonType, "--format", "json").Output() //nolint:gosec // args from operator flags
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			slog.Error("ceph orch ps", "daemon_type", daemonType, "error", string(exitErr.Stderr))
		} else {
			slog.Error("ceph orch ps", "daemon_type", daemonType, "error", err)
		}
		return nil, err
	}

	daemons, err := parseOrchPS(b)
	if err != nil {
		slog.Error("failed to unmarshal JSON", "error", err)
		return nil, err
	}

	return daemons, nil
}

// parseOrchPS unmarshals the JSON array output of `ceph orch ps --format json` into a slice of [DaemonStatus]
//
// Daemons with an empty status_desc are reported as "unknown", so that every daemon is counted under some status.
func parseOrchPS(b []byte) ([]DaemonStatus, error) {
	var daemons []DaemonStatus
	if err := json.Unmarshal(b, &daemons); err != nil {
		return nil, err
	}

	for i := range daemons {
		if daemons[i].StatusDesc == "" {
			daemons[i].StatusDesc = "unknown"
		}
	}

	return daemons, nil
}
