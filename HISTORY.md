# History: tenantecho

Append-only: add entries at the end, never edit or remove one (CI enforces it). One line each:
`- <YYYY-MM-DD> <who: looper|agent|jody> <kind: issue|decision|release|deploy> <one line> (<links>)`

- 2026-11-03 looper issue PingPigeon drops the inbound trace_id on CreatePigeon, the task publish and CancelPigeon, with no test for it (#8, b173d52)
