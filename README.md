# rbd-exporter

This is an _experimental_ prometheus exporter to get some stats for Ceph RBD that don't seem to be available elsewhere.
It exports:

| Metric | Source |
| ------ | ------ |
| `rbd_mirror_pool_status_state{pool,state}` | `rbd mirror pool status <pool>` |
| `rbd_mirror_leader_watchers{pool}` | `rados -p <pool> listwatchers rbd_mirror_leader` |
| `cephadm_daemon_status{daemon_type,status}` | `ceph orch ps --daemon-type <daemon>` |

See <https://docs.ceph.com/en/reef/rbd/rbd-mirroring/#mirror-status> for more info on what the pool status means.

A healthy mirror normally has as many leader watchers as running rbd-mirror daemons.

The capabilities of this exporter *might* grow over time.
