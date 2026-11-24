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

# The release scripts' tests (ci/release_test.sh): the image tags, notes, tacplus.lock and justfile a
# release takes, with git and just alone. From this justfile's directory, not the working directory: a
# release lends this justfile to a tag older than it (v0.1.0), whose source has no ci/ (ci/release-build.sh).
release-test:
  {{justfile_directory()}}/ci/release_test.sh

# Every FROM of the Dockerfiles pinned by digest, never a floating tag (ci/dockerfile_test.sh, #15).
dockerfile-test:
  {{justfile_directory()}}/ci/dockerfile_test.sh

# The ping page (web/, README), as the fleet builder's webapp recipe builds it: its client generated, its
# tests, and its build twice to the same bytes (ci/web_test.sh), against the working directory's source.
web:
  TACPLUS_BIN={{TACPLUS}} {{justfile_directory()}}/ci/web_test.sh

# The gates every change passes before it lands: the build job runs them in the pinned build image, and
# release.yml runs them on the tag it releases (README, Releasing).
gates: test release-test dockerfile-test web

clean:
  rm -rf tacplus tenantecho web/model web/node_modules web/dist
