# State: tenantecho

Overwritten to stay current (loopers and agents). Last updated: 2026-11-06 by looper (#5).

- **Version / line:** unreleased (no tag yet, #2); tacplus floor 2.5
- **Works:** `TenantEcho.Ping` and `PingPigeon`, each routed by its request's `tenant` (the target tid, `route_field`), default tenant door: a call reaches the named tenant's instance, another tenant's instance refuses it (`PERMISSION_DENIED`), an empty or malformed tid is refused (`INVALID_ARGUMENT`), tested against the generated wrapper and authorizer (#5). `PingPigeon` carries its request's `trace_id` on `CreatePigeon`, the task message and `CancelPigeon` (#8). Generation, build, vet and tests pass with plain `tacplus` 2.5.11 (its compiler image now generates Go).
- **Released / deployed:** nowhere from this repository yet; fleet-seatrials builds it from its own proto copy (fleet on tacplus 2.2).
- **Known gaps:** fleet-seatrials' client does not fill `tenant` yet (PROPOSALS.md). `looper-trace check` sees none of the handlers (PROPOSALS.md).
- **Blocked on:** #2 on tacplus-dev/looper#12.
