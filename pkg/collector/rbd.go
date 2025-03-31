package collector

import (
	"encoding/json"
	"log/slog"
	"os/exec"
)

// RBDMirrorPoolStatus implements the [PoolStatusProvider] interface to exec the `rbd mirror pool status` command
type RBDMirrorPoolStatus struct{}

// GetPoolStatus implements the [PoolStatusProvider] interface
//
// This could be implemented by some kind of wrapper around librbd, but
// for now we'll just exec out.
func (r *RBDMirrorPoolStatus) GetPoolStatus(pool string) (PoolStatus, error) {
	// Execute the rbd command to get the pool status
	// The command should be run with the appropriate arguments and options
	// to retrieve the desired information
	// For example: rbd mirror pool status --format json <pool>

	// Execute the command and capture the output

	b, err := exec.Command("rbd", "mirror", "pool", "status", "--format", "json", pool).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			slog.Error("rbd mirror pool status", "pool", pool, "error", string(exitErr.Stderr))
		} else {
			slog.Error("rbd mirror pool status", "pool", pool, "error", err)
		}
		return PoolStatus{}, err
	}

	// Unmarshal the JSON output into the poolStatus struct
	var s PoolStatus
	err = json.Unmarshal(b, &s)
	if err != nil {
		slog.Error("failed to unmarshal JSON", "error", err)
		return PoolStatus{}, err
	}

	return s, nil
}
