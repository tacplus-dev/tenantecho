# State: tenantecho

Overwritten to stay current (loopers and agents). Last updated: 2026-10-29 by looper (#5).

- **Version / line:** unreleased (no tag yet, #2); tacplus floor 2.5
- **Works:** `TenantEcho.Ping` and `PingPigeon`, each routed by its request's `tenant` (the target tid, `route_field`), default tenant door. Generation, build, vet and tests pass on tacplus 2.5.1 (and 2.5.0) when the compiler image has a Go toolchain on PATH.
- **Released / deployed:** nowhere from this repository yet; fleet-seatrials builds it from its own proto copy (fleet on tacplus 2.2).
- **Known gaps:** plain `tacplus` in the looper cannot generate Go (the `tacplus:2.5` compiler image, 2.5.1 included, has no `go`), so the gates run only with a mounted toolchain. The fix (tacplus-dev/tacplus#193, the build image) ships first in 2.6.0, with no `tacplus` image, so a 2.5 floor never gets it (PROPOSALS.md). fleet-seatrials' client does not fill `tenant` yet (PROPOSALS.md).
- **Blocked on:** tacplus-dev/tacplus#213 (v2.6.0, the first release with the build image from #193) and tacplus-dev/tacplus#188 (no `go` to build a Go service), for the gates of #1 and #5; #2 on tacplus-dev/looper#12.
