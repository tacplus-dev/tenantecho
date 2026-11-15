#!/usr/bin/env bash
# The gates and the binaries of a release (.github/workflows/release.yml), in the build image the release's
# tacplus.lock pins by digest: `just gates`, as the build job runs them on every land, with the image's own
# tacplus, then the static tenantecho for each architecture, to SRC/bin/linux-<arch>/tenantecho.
#
# Usage:  ci/release-build.sh SRC
#
# SRC is the release's source, a scratch checkout of its tag. The lock and the justfile are SRC's own, or
# the workflow's commit's for a tag cut before this repository had them (ci/release-file.sh; v0.1.0 has
# neither): a borrowed lock is copied in, since tacplus reads it from the source; a borrowed justfile runs
# against SRC (--working-directory) and its release tests from its own ci/, so the workflow's checkout is
# mounted too. That mount also gives SRC its git when it is a worktree of the workflow's checkout, as the
# gates need (justfile, generate). ARCHS (default "arm64") are the architectures to build. Go's VCS stamping
# is off (GOFLAGS): the binary's commit is the image's revision label (ci/release-image.sh), not a stamp.
set -euo pipefail

src="$(cd "$1" && pwd)"
ci="$(cd "$(dirname "$0")" && pwd)"
root="$(cd "$ci/.." && pwd)"
lock="$("$ci/release-file.sh" "$src" tacplus.lock)"
justfile="$("$ci/release-file.sh" "$src" justfile)"
image="$(sed -n -E 's/^build_image:[[:space:]]*"?([^"[:space:]]+)"?.*$/\1/p' "$lock" | head -1)"
[[ "$image" == *@sha256:* ]] || { echo "release-build.sh: $lock names no build_image by digest" >&2; exit 1; }
[[ "$lock" == "$src/tacplus.lock" ]] || cp "$lock" "$src/tacplus.lock"
echo "release-build.sh: tacplus $(sed -n 's/^tacplus:[[:space:]]*//p' "$lock"), gates of $justfile, in $image"

docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -e GOFLAGS=-buildvcs=false -e TACPLUS_BIN=tacplus \
  -e ARCHS="${ARCHS:-arm64}" -e JUSTFILE="$justfile" -v "$src:$src" -v "$root:$root" -w "$src" \
  "$image" sh -euc '
    just --justfile "$JUSTFILE" --working-directory "$PWD" gates
    for arch in $ARCHS; do
      CGO_ENABLED=0 GOOS=linux GOARCH=$arch go build -trimpath -buildvcs=false -ldflags="-s -w -buildid=" \
        -o "bin/linux-$arch/tenantecho" ./cmd/tenantecho
    done'
