# Copyright (c) 2026 Jody Scott
# SPDX-License-Identifier: Apache-2.0
#
# `just gates`: the checks every change to main passes, which the build pipeline
# (tacplus-dev/build) runs in the build image tacplus.lock pins, by digest. In that
# image `tacplus` runs exactly the pinned release; TACPLUS_BIN overrides it.
TACPLUS := env_var_or_default("TACPLUS_BIN", "tacplus")

default: gates

# Regenerate tacplus/ (Go) and web/model/ (TypeScript), neither committed, with the
# pinned tacplus. Generated code is never committed or edited by hand, so
# generation must leave every tracked file as it found it.
generate:
  #!/usr/bin/env bash
  set -euo pipefail
  before="$(git status --porcelain --untracked-files=no)"
  {{TACPLUS}} build
  (cd web && {{TACPLUS}} build)
  if [[ "$(git status --porcelain --untracked-files=no)" != "$before" ]]; then
    echo "generate: tacplus build changed tracked files:" >&2; git status --short >&2; exit 1
  fi
  if [[ -n "$(git ls-files tacplus web/model)" ]]; then
    echo "generate: tacplus/ and web/model/ are generated and must not be committed" >&2; exit 1
  fi

# The service: build, vet and test against what generate just wrote.
go: generate
  go build ./...
  go vet ./...
  go test ./...

# The ping page: its tests, and the build the fleet's gateway serves (web/dist/).
web: generate
  cd web && npm ci --ignore-scripts
  cd web && npx vitest run
  cd web && TACPLUS_WEB_BASE=/-/apps/tenantecho/ npm run build

# The gates every change passes before it lands.
gates: go web

clean:
  rm -rf tacplus web/model web/dist web/node_modules
