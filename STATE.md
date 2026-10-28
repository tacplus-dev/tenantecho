# State: tenantecho

Overwritten to stay current (loopers and agents). Last updated: 2026-10-28 by looper (#5).

- **Version / line:** unreleased (no tag yet, #2); tacplus floor 2.5
- **Works:** `TenantEcho.Ping` and `PingPigeon`, each routed by its request's `tenant` (the target tid, `route_field`), default tenant door. Generation, build, vet and tests pass on tacplus 2.5.0 when the compiler image has a Go toolchain on PATH.
- **Released / deployed:** nowhere from this repository yet; fleet-seatrials builds it from its own proto copy (fleet on tacplus 2.2).
- **Known gaps:** plain `tacplus` in the looper cannot generate Go (the compiler image has no `go`), so the gates run only with a mounted toolchain. fleet-seatrials' client does not fill `tenant` yet (PROPOSALS.md).
- **Blocked on:** tacplus-dev/tacplus#193 (the build image that generation runs in, with Go), for the gates of #1 and #5; #2 on tacplus-dev/looper#12.
