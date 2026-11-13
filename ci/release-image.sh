#!/usr/bin/env bash
# Assembles the release image (.github/workflows/release.yml) from SRC's Dockerfile, once per architecture,
# and publishes them as one multi-architecture image.
#
# Usage:  ci/release-image.sh VERSION REVISION SRC LAUNCHERS [TAG...]
#
# The Dockerfile takes its binaries from beside it, as the fleet's image pipeline puts them there: for each
# architecture this stages a build context with SRC/bin/linux-<arch>/tenantecho (ci/release-build.sh) as
# bin/tenantecho and LAUNCHERS/launcher-linux-<arch>, and builds it labelled with VERSION and REVISION.
# With TAGs, each architecture is pushed by digest and $IMAGE:<TAG> (each TAG) is pointed at all of them.
# Without, nothing is pushed: the dry run. ARCHS (default "arm64") are the architectures.
set -euo pipefail

version="$1"
revision="$2"
src="$(cd "$3" && pwd)"
launchers="$(cd "$4" && pwd)"
shift 4
image="${IMAGE:-ghcr.io/tacplus-dev/tenantecho}"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

digests=()
for arch in ${ARCHS:-arm64}; do
  case "$arch" in
    amd64|arm64) ;;
    *) echo "release-image.sh: unsupported architecture $arch" >&2; exit 1 ;;
  esac
  ctx="$work/$arch"
  mkdir -p "$ctx/bin"
  cp "$src/Dockerfile" "$ctx/"
  cp "$src/bin/linux-$arch/tenantecho" "$ctx/bin/tenantecho"
  cp "$launchers/launcher-linux-$arch" "$ctx/"

  output="type=cacheonly"
  (($#)) && output="type=image,name=$image,push-by-digest=true,name-canonical=true,push=true"
  docker buildx build --platform "linux/$arch" \
    --build-arg "VERSION=$version" --build-arg "VCS_REF=$revision" \
    --output "$output" --metadata-file "$ctx.json" "$ctx"
  if (($#)); then
    digest="$(jq -r '."containerimage.digest"' "$ctx.json")"
    echo "linux/$arch: $image@$digest"
    digests+=("$image@$digest")
  fi
done

(($#)) || exit 0
refs=()
for tag in "$@"; do refs+=(-t "$image:$tag"); done
docker buildx imagetools create "${refs[@]}" "${digests[@]}"
