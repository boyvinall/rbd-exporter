package collector

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

// RadosListWatchers implements the [WatcherProvider] interface to exec the `rados listwatchers` command
type RadosListWatchers struct{}

// GetWatchers implements the [WatcherProvider] interface
//
// This could be implemented by some kind of wrapper around librados, but
// for now we'll just exec out.
func (r *RadosListWatchers) GetWatchers(pool, object string) ([]Watcher, error) {
	// Execute the command and capture the output
	// For example: rados -p <pool> listwatchers rbd_mirror_leader

	b, err := exec.Command("rados", "-p", pool, "listwatchers", object).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			slog.Error("rados listwatchers", "pool", pool, "object", object, "error", string(exitErr.Stderr))
		} else {
			slog.Error("rados listwatchers", "pool", pool, "object", object, "error", err)
		}
		return nil, err
	}

	// The command has no JSON output format, so parse the text output
	watchers, err := parseListWatchers(b)
	if err != nil {
		slog.Error("failed to parse rados listwatchers output", "pool", pool, "object", object, "error", err)
		return nil, err
	}

	return watchers, nil
}

// parseListWatchers parses the text output of the `rados listwatchers` command
//
// Each non-blank line describes one watcher, with whitespace-separated fields, e.g.
//
//	watcher=192.0.2.1:0/9029529117 client.2912896056 cookie=59662829605819
//
// Blank lines are skipped, and empty output means there are no watchers.
func parseListWatchers(b []byte) ([]Watcher, error) {
	watchers := []Watcher{}
	for _, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		if len(fields) != 3 {
			return nil, fmt.Errorf("unexpected rados listwatchers line: %q", line)
		}

		address, okAddress := strings.CutPrefix(fields[0], "watcher=")
		cookie, okCookie := strings.CutPrefix(fields[2], "cookie=")
		if !okAddress || !okCookie || !strings.HasPrefix(fields[1], "client.") || address == "" || cookie == "" || fields[1] == "client." {
			return nil, fmt.Errorf("unexpected rados listwatchers line: %q", line)
		}

		watchers = append(watchers, Watcher{
			Address: address,
			Client:  fields[1],
			Cookie:  cookie,
		})
	}

	return watchers, nil
}
