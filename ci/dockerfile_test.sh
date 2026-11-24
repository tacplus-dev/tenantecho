#!/usr/bin/env bash
# Every FROM of the repository's Dockerfiles is pinned by digest, <image>:<tag>@sha256:<digest> (#15;
# tacplus-dev/roadmap decisions/security-gates-ship-only-2026-11-23.md, point 2: immutable tags only, never a
# floating one), so a rebuild of the same commit ships the same, scanned base. A FROM naming an earlier stage
# of the same Dockerfile is not an image and needs none. A FROM through a build argument (FROM ${BASE_IMAGE})
# is refused: its default may be pinned, but a --build-arg moves it. Nightly maintenance moves the digests
# (tacplus-dev/looper#234).
#
# Usage:  ci/dockerfile_test.sh
set -euo pipefail

ci="$(cd "$(dirname "$0")" && pwd)"
scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT
failed=0

# Prints each FROM of a Dockerfile that is not pinned by digest, "<file>:<line>: <image>"; fails if any.
unpinned() { # dockerfile
  awk -v file="$1" '
    toupper($1) == "FROM" {
      i = 2
      while ($i ~ /^--/) i++
      image = $i
      if (toupper($(i + 1)) == "AS") stages[tolower($(i + 2))] = 1
      if (tolower(image) in stages) next
      if (image ~ /\$/ || image !~ /@sha256:[0-9a-f]{64}$/) { print file ":" NR ": " image; bad = 1 }
    }
    END { exit bad }
  ' "$1"
}

check() { # name dockerfile-content want(pinned|unpinned)
  printf '%s\n' "$2" > "$scratch/Dockerfile"
  local got=pinned
  unpinned "$scratch/Dockerfile" > /dev/null || got=unpinned
  if [[ "$got" == "$3" ]]; then
    echo "ok   $1"
  else
    echo "FAIL $1: got $got, want $3"
    failed=1
  fi
}

digest=sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf
check "by digest" "FROM alpine:3.24@$digest" pinned
check "lowercase from, by digest" "from alpine:3.24@$digest" pinned
check "a floating tag" "FROM alpine:3.24" unpinned
check "no tag" "FROM alpine" unpinned
check "a short digest" "FROM alpine:3.24@sha256:f59abe65" unpinned
check "a build argument, its default pinned" "$(printf 'ARG BASE_IMAGE=alpine:3.24@%s\nFROM ${BASE_IMAGE}' "$digest")" unpinned
check "--platform, by digest" "FROM --platform=\$BUILDPLATFORM golang:1.27@$digest AS build" pinned
check "--platform, a floating tag" "FROM --platform=\$BUILDPLATFORM golang:1.27 AS build" unpinned
check "an earlier stage" "$(printf 'FROM golang:1.27@%s AS build\nFROM build' "$digest")" pinned
check "a builder stage unpinned" "$(printf 'FROM golang:1.27 AS build\nFROM alpine:3.24@%s' "$digest")" unpinned
check "a later stage's name is not an earlier stage" "$(printf 'FROM build\nFROM alpine:3.24@%s AS build' "$digest")" unpinned

# The repository's own Dockerfiles, tracked or not yet.
root="$(cd "$ci/.." && pwd)"
found=0
while IFS= read -r f; do
  found=1
  if unpinned "$root/$f" > "$scratch/out"; then
    echo "ok   $f: every FROM by digest"
  else
    echo "FAIL $f: FROM not pinned by digest:"; sed 's/^/       /' "$scratch/out"
    failed=1
  fi
done < <(git -C "$root" ls-files --cached --others --exclude-standard -- '*Dockerfile*')
((found)) || { echo "FAIL no Dockerfile found under $root"; failed=1; }

exit "$failed"
