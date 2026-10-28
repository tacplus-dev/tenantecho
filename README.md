# tenantecho

The smallest tenant-owned tacplus service. `(spec.service).tenant_owned` makes the operator run
one instance per tenant namespace, each with the identity
`spiffe://<td>/tenants/<tenant>/services/tenantecho` and its own tenant's `Config`.

`TenantEcho.Ping` (client-facing through the gateway, any authenticated user) answers with the
message, `served_by` — the answering instance's own identity path, read from its own token at
start — the caller's tenant namespace from the caller's token, and the instance's configured
`greeting`. A call that the gateway routed to the wrong tenant's instance says so in its own
response: `served_by` names a tenant other than `caller_namespace`.

Every rpc routes by its request's `tenant`, the target tenant's tid (tacplus 2.5 `route_field`): the
caller names it, never the token, and an empty or malformed one is refused (`INVALID_ARGUMENT`).
The door is the default one, without `RPC_FLAG_ANY_TENANT`: only members of the target tenant get
in. So a call lands in the caller's own tenant's instance or is refused, never another tenant's.

That is the whole purpose: deployed into two tenant namespaces by the proof fleet
(`fleet-seatrials`), it shows the gateway routes each tenant's call to that tenant's instance.

Registered in the tacplus registry as `tenantecho`, id 103.
