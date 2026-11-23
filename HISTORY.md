# History: tenantecho

Append-only: add entries at the end, never edit or remove one (CI enforces it). One line each:
`- <YYYY-MM-DD> <who: looper|agent|jody> <kind: issue|decision|release|deploy> <one line> (<links>)`

- 2026-11-03 looper issue PingPigeon drops the inbound trace_id on CreatePigeon, the task publish and CancelPigeon, with no test for it (#8, b173d52)
- 2026-11-06 looper issue Declare the tacplus floor: 2.4 (#1)
- 2026-11-06 looper issue Bump golang.org/x/net to v0.61.0: five HTTP/2 advisories (GO-2026-6603/6610/6611/6612/6617) (#3)
- 2026-11-06 looper issue Bump golang.org/x/net to v0.62.0: GO-2026-6860 and GO-2026-6866 in x/net/html (#6)
- 2026-11-06 looper issue Base image: alpine:3.22 is not the latest Alpine line (3.24) (#4)
- 2026-11-09 looper issue tacplus floor 2.5 is behind the latest released line (2.6): move the floor to 2.6 and commit tacplus.lock (#9)
- 2026-11-15 looper issue Release: tag vMAJOR.MINOR.PATCH (start at v0.1.0) and publish its image (#2, e271a88)
- 2026-11-15 looper issue Declare route fields: a target tenant on every tenant-owned rpc, floor 2.5 (#5)
- 2026-11-15 looper release v0.2.0 (https://github.com/tacplus-dev/tenantecho/releases/tag/v0.2.0)
- 2026-11-15 looper issue Contribute a ping page to the fleet's gateway: the example module (tacplus#195) (#7, a59708b)
- 2026-11-15 looper release v0.2.1 (https://github.com/tacplus-dev/tenantecho/releases/tag/v0.2.1)
- 2026-11-15 looper issue GO-2026-6859 (net/http HTTP/1 crash): the pinned build image (tacplus 2.6.1) compiles with go1.27.3; re-pin to 2.6.8 (go1.27.4) (#11, 67286ce)
- 2026-11-15 looper issue Pin the build: floor 2.6 and a committed tacplus.lock, so the build pipeline and its security checks run on this repository (#10, 91359ef)
- 2026-11-17 looper issue Security checks: new non-blocking trivy DS-0026 (LOW), the release Dockerfile has no HEALTHCHECK: fix or except (#12)
- 2026-11-23 looper release v0.2.3 (https://github.com/tacplus-dev/tenantecho/releases/tag/v0.2.3)
