# Proposals: tenantecho

Written by loopers. Each proposal: what, why, links. Jody promotes accepted ones into ROADMAP.md (via an agent);
the looper then marks it `promoted` or `declined` here with the date.

## Open
- **<title>** (<date>, <issue/run link>): <what and why>
- **fleet-seatrials: name the target tenant in tenantecho calls** (2026-10-28, #5): since #5, `PingRequest.tenant` (2) and `PingPigeonRequest.tenant` (3) are the route fields, and a call with an empty or malformed one is refused with `INVALID_ARGUMENT` before it is sent. The seatrials client sends neither (tacplus-dev/fleet-seatrials `client/cmd/seatrials/main.go`, `ping` and `pingPigeon`). When that fleet moves to tacplus 2.5, it should send the credential's tenant tid, and maybe add a `-tenant` flag to show that a call naming another tenant is refused (`PERMISSION_DENIED`). Its `build/proto/services/tenantecho.proto` copy also has to follow this one.
- **README examples use tenant names, not tids** (2026-10-28, #5): the README and tests show `tenants/acme`. Since 2.5, a tenant's namespace is `/tenants/<tid>` (8 characters of `[0-9a-z]`), and the route field must be a tid. The examples should use a tid, so they match what a caller now sends.

## Promoted / declined
