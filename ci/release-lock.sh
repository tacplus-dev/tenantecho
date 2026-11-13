#!/usr/bin/env bash
# The tacplus.lock a release of SRC builds with (ci/release-build.sh): SRC's own, or, for a tag cut before
# this repository pinned one (v0.1.0), the lock beside these scripts, the workflow's commit's, saying so on
# stderr. Its pin must still satisfy SRC's floor: the build image's tacplus refuses it otherwise.
#
# Usage:  ci/release-lock.sh SRC     prints the lock's path
set -euo pipefail

src="$(cd "$1" && pwd)"
if [[ -f "$src/tacplus.lock" ]]; then
  echo "$src/tacplus.lock"
  exit 0
fi
own="$(cd "$(dirname "$0")/.." && pwd)/tacplus.lock"
[[ -f "$own" ]] || { echo "release-lock.sh: neither $src nor the workflow's commit has a tacplus.lock" >&2; exit 1; }
echo "release-lock.sh: $src has no tacplus.lock: building with the workflow's commit's ($(sed -n 's/^tacplus:[[:space:]]*//p' "$own"))" >&2
echo "$own"
