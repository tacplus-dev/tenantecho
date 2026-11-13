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

## The ping page (`web/`)

tenantecho is the example contribution to a fleet's gateway (tacplus `docs/USER.md`, "Contributing
to the gateway"): `web/` is a small webapp module, served by the gateway at `/-/apps/tenantecho/`. It
pings the tenant you name through its generated TypeScript W2S client and shows the reply and the
tenant whose instance answered. It signs nobody in: it uses the gateway's session on the same origin
and sends a signed-out user to `/`.

`web/tacplus.yaml` generates the client into `web/model/` (not committed; the fleet builder
regenerates it against the fleet's merged protos). To build and test it:

```sh
cd web
tacplus build
npm ci --ignore-scripts
npx vitest run
TACPLUS_WEB_BASE=/-/apps/tenantecho/ npm run build   # writes web/dist/
```

## Releasing

tenantecho is released by version, `vMAJOR.MINOR.PATCH`, so fleets pin it by release line
(tacplus-dev/tacplus#180). It stays on `0.x` until SeaBeam's launch, when `1.0.0` is cut; on `0.x` a
minor may break, a patch never does. A fix is a patch, a feature a minor; both are tagged on main, and a
patch to an older line on its `release/MAJOR.MINOR` branch.

To cut a release, push an annotated tag on a green commit:

```sh
git tag -a v0.2.0 -m "tenantecho 0.2.0: <what it brings>"
git push origin v0.2.0
```

The tag runs `.github/workflows/release.yml`: the gates on the tag (generate, build, vet, test) in the
build image `tacplus.lock` pins, the image `ghcr.io/tacplus-dev/tenantecho:0.2.0` labelled with its
version and commit, and the line tag `0.2`, which moves only to the line's highest patch, never
backwards nor on another line's release. Then the GitHub release, with the commits since the previous
release as its notes; a patch on an older line is never marked Latest. The image is linux/arm64; amd64 is
off by default, as in the fleets, and `ARCHS: amd64 arm64` in the workflow builds both. A tag is never
moved or deleted.

The workflow can also be run by hand (Actions, release, Run workflow) with a version:

- without `publish`, a dry run: it builds the tag, or main tagged locally as that version, and prints the
  image tags and notes that release would get, publishing nothing;
- with `publish`, it releases an existing tag that has no release yet, as its tag push would have, with the
  scripts of the workflow's commit: for a tag cut before the workflow existed (v0.1.0). A tag without a
  `tacplus.lock` builds with that commit's pin, and its notes say so.

The package must be public, so pulls need no credential: the first release's last step fails, naming it,
until an org owner makes it public once (the package's settings, Change visibility). The release scripts'
tests: `ci/release_test.sh`.
