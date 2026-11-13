#!/usr/bin/env bash
# The gates and the binaries of a release (.github/workflows/release.yml), in the build image the release's
# tacplus.lock pins by digest (ci/release-lock.sh): generate, build, vet and test as main's gates do, then
# the static tenantecho for each architecture, to SRC/bin/linux-<arch>/tenantecho.
#
# Usage:  ci/release-build.sh SRC
#
# SRC is the release's source, a scratch checkout of its tag: a lock taken from the workflow's commit is
# copied into it. ARCHS (default "arm64") are the architectures to build.
set -euo pipefail

src="$(cd "$1" && pwd)"
lock="$("$(dirname "$0")/release-lock.sh" "$src")"
image="$(sed -n -E 's/^build_image:[[:space:]]*"?([^"[:space:]]+)"?.*$/\1/p' "$lock" | head -1)"
[[ "$image" == *@sha256:* ]] || { echo "release-build.sh: $lock names no build_image by digest" >&2; exit 1; }
[[ "$lock" == "$src/tacplus.lock" ]] || cp "$lock" "$src/tacplus.lock"

docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -e ARCHS="${ARCHS:-arm64}" -v "$src:$src" -w "$src" \
  "$image" sh -euc '
    tacplus build
    go build ./...
    go vet ./...
    go test ./...
    for arch in $ARCHS; do
      CGO_ENABLED=0 GOOS=linux GOARCH=$arch go build -trimpath -buildvcs=false -ldflags="-s -w -buildid=" \
        -o "bin/linux-$arch/tenantecho" ./cmd/tenantecho
    done'
