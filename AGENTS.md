# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

An experimental Prometheus exporter for the health of Ceph RBD mirroring. It shells out to the `rbd`, `rados` and `ceph` CLIs, so it needs `ceph-common` plus a Ceph config and keyring at runtime. `ceph orch` also needs a reachable mgr with the cephadm orchestrator. The Docker image provides these, and `cmd/rbd-exporter/spec.yml` deploys it as a cephadm custom container on mgr-labelled hosts.

## Commands

```sh
make build          # cross-compile (CGO_ENABLED=0) to out/{linux-amd64,linux-arm64,darwin-amd64}/rbd-exporter
make test           # go test via `go tool gocov`, writes coverage.html
make lint           # golangci-lint run (CI pins v1.64.5; .golangci.yml uses the v1 config format)
make docker-build   # builds out/linux-amd64 binary, then image boyvinall/rbd-exporter:$(TAG)

go test ./pkg/collector -run TestCollector_collectPoolMetrics/return-all-data   # single test case
go run ./cmd/rbd-exporter -v serve --pool <pool> [--pool <pool2>] [--listen-address :9876]
```

`gocov` and `gocov-html` are declared as Go tool dependencies in `go.mod`, so `go tool` resolves them and you don't install them separately.

A golangci-lint v2 install (e.g. the current Homebrew version) refuses the v1-format `.golangci.yml` with "unsupported version of the configuration". To lint exactly as CI does, run `go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5 run`.

## Architecture

- `cmd/rbd-exporter/main.go`: urfave/cli v2 app with a single `serve` command. It registers one `collector.Collector` with the default Prometheus registry and serves `/metrics` on `:9876` by default.
- `pkg/collector/collector.go`: `Collector` implements `prometheus.Collector` and gathers data synchronously on every scrape, with no caching. Each data source sits behind a provider interface, so `collector.go` never names a concrete provider or Ceph command. `main.go` wires in the real exec-based providers, and tests swap in mocks.

| `collect*` func | Interface → implementation | Command | Metric |
| --- | --- | --- | --- |
| `collectPoolMetrics(pool)` | `PoolStatusProvider` → `RBDMirrorPoolStatus` (`rbd.go`) | `rbd mirror pool status --format json <pool>` | `rbd_mirror_pool_status_state{pool,state}` |
| `collectLeaderWatcherMetrics(pool)` | `WatcherProvider` → `RadosListWatchers` (`rados.go`) | `rados -p <pool> listwatchers rbd_mirror_leader` | `rbd_mirror_leader_watchers{pool}` |
| `collectDaemonMetrics(daemonType)` | `DaemonStatusProvider` → `CephOrchPS` (`orch.go`) | `ceph orch ps --daemon-type <daemonType> --format json` | `cephadm_daemon_status{daemon_type,status}` |

- **What `Collect` runs:** sources that take a pool argument run inside the loop over the `--pool` list; daemon sources run inside a loop over the `--daemon` list (defaulting to `rbd-mirror` when none is given). Each runs its command once per argument, labelling the series with that argument so they don't collide; emitting duplicate series makes Prometheus reject the scrape.
- **Zero-filling:** known states and statuses are always reported, as zero when absent, and unknown values from Ceph are passed through.
- **Errors:** each `collect*` result goes through the `send` closure, which turns an error into an invalid `rbd_exporter_error` metric, so sources fail independently. However, `promhttp.Handler()` uses default options, so any invalid metric makes the whole `/metrics` response an HTTP 500.

To add a metric source, add an interface and its data type to `collector.go`, put the exec implementation in its own file (modelled on `rbd.go`), and keep the parsing in a pure helper (`parseOrchPS`, `parseListWatchers`) so tests don't need the Ceph binaries. Then call the new function from `Collect` (inside or outside the pool loop) and wire the provider into `collector.New` in `main.go`. Test fixtures are real cluster output (Ceph 18.2.7), anonymized: hostnames are `serverN`, IPs are in `192.0.2.0/24`, and IDs are random. Use real output for new fixtures too, rather than inventing it, and anonymize it the same way.

## Conventions

- Commit messages and PR titles must follow Conventional Commits (enforced by commitlint in CI). semantic-release on `main` uses them to cut versions. When a release is created, CI runs `make release` (docker push with the git tag) instead of build and test.
- Logging uses `log/slog`. The `-v` global flag (placed before `serve`) turns on debug level.
