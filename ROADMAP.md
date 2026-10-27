# Roadmap: tenantecho

Owned by Jody; changed only by an agent at Jody's request. Loopers propose in PROPOSALS.md.

## Now
The smallest tenant-owned tacplus service: `TenantEcho.Ping` proves the gateway routes each tenant's call to that tenant's instance (fleet-seatrials).
Go. Unreleased; tacplus floor 2.4 (declared, #1 still open).

## Next
- Nothing approved and unblocked; every open approved issue is in Later.

## Later
- Declare the tacplus floor 2.4 and pass the gates on it (#1), blocked on tacplus-dev/tacplus#189
- Release v0.1.0 and publish its image (#2), blocked on tacplus-dev/looper#12 and #1
- golang.org/x/net to v0.61.0 for five HTTP/2 advisories (#3), blocked on #1
- Base image alpine:3.22 to alpine:3.24 (#4), blocked on #1

## Not doing
- Nothing recorded yet.
