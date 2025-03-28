# rbd-exporter

This is an _experimental_ prometheus exporter to get some stats for Ceph RBD that don't seem to be available elsewhere.
Specifically, looking to get the output of the following command into prometheus:

```plaintext
rbd mirror pool status <pool>
```

See <https://docs.ceph.com/en/reef/rbd/rbd-mirroring/#mirror-status> for more info on what this does.

The capabilities of this exporter *might* grow over time.
