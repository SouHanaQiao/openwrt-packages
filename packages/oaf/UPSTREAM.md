# OAF kernel module source

- Upstream: https://github.com/destan19/OpenAppFilter
- Version/tag: `v6.1.3`
- Commit: `29142c9d0985c83572b627fe5978ea0590ef0cbe`
- Imported: 2026-08-22

This module is intentionally kept at v6.1.3 to match the existing
`appfilter` 6.1.3 userspace package in this feed. It must be rebuilt by
OpenWrt for the exact kernel ABI used by the target firmware.

The feed enables module autoload so `/proc/sys/oaf` exists before the
`appfilter` service starts after boot.
