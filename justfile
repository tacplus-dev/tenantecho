# Copyright (c) 2026 Jody Scott
# SPDX-License-Identifier: Apache-2.0

# The compiler: in the build image tacplus.lock pins, the image's own (the build job and the release set
# TACPLUS_BIN=tacplus); TACPLUS_BIN overrides it with a tacplus of your own.
TACPLUS := env_var_or_default("TACPLUS_BIN", "tacplus")

default: gates

# Regenerate tacplus/ (not committed) with the compiler. Generated code is never committed or edited by
# hand, so generation must leave every tracked file as it found it.
generate:
  #!/usr/bin/env bash
  set -euo pipefail
  if [[ -n "$(git ls-files tacplus)" ]]; then
    echo "generate: tacplus/ is generated and must not be committed" >&2; exit 1
  fi
  before="$(git diff HEAD | sha256sum)"
  {{TACPLUS}} build
  if [[ "$(git diff HEAD | sha256sum)" != "$before" ]]; then
    echo "generate: tacplus build changed tracked files:" >&2; git status --short >&2; exit 1
  fi

build: generate
  go build ./...
  go vet ./...

test: build
  go test ./...

# The release scripts' tests (ci/release_test.sh): the image tags, notes and tacplus.lock a release takes,
# with git alone.
release-test:
  ci/release_test.sh

# The gates every change passes before it lands: the build job runs them in the pinned build image, and
# release.yml runs them on the tag it releases (README, Releasing).
gates: test release-test

clean:
  rm -rf tacplus tenantecho
