# State: tenantecho

Overwritten to stay current (loopers and agents). Last updated: 2026-11-07 by looper (#7).

- **Version / line:** v0.1.0 tagged (source only: no GitHub release or image yet, #2); tacplus floor 2.5
- **Works:** `TenantEcho.Ping` and `PingPigeon`, each routed by its request's `tenant` (the target tid, `route_field`), default tenant door: a call reaches the named tenant's instance, another tenant's instance refuses it (`PERMISSION_DENIED`), an empty or malformed tid is refused (`INVALID_ARGUMENT`), tested against the generated wrapper and authorizer (#5). `PingPigeon` carries its request's `trace_id` on `CreatePigeon`, the task message and `CancelPigeon` (#8). Generation, build, vet and tests pass on the 2.6 line; on the 2.5 floor generation now fails, because `coop`'s latest release needs 2.6 (#9). `web/` is the ping page tenantecho contributes to its fleet's gateway (#7): a TypeScript W2S client (`web/tacplus.yaml`, floor 2.6) and a Vite page, tested with vitest, whose `npm run build` with `TACPLUS_WEB_BASE=/-/apps/tenantecho/` writes the same `web/dist/` bytes every time.
- **Released / deployed:** nowhere from this repository yet; fleet-seatrials builds it from its own proto copy (fleet on tacplus 2.2).
- **Known gaps:** the root `tacplus.yaml` does not declare `contribute.webapp` yet: that waits for the 2.6 floor (#9), then a release carries it (#7). fleet-seatrials' client does not fill `tenant` yet (PROPOSALS.md). `looper-trace check` sees none of the handlers (PROPOSALS.md).
- **Blocked on:** #2 (its GitHub release and image) on tacplus-dev/looper#17. #7's `contribute:` on #9.
