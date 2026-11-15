#!/usr/bin/env bash
# The FILE a release of SRC builds with (ci/release-build.sh): SRC's own, or, for a tag cut before this
# repository had one (v0.1.0 has neither tacplus.lock nor justfile), the one beside these scripts, the
# workflow's commit's, saying so on stderr. The lock's pin must still satisfy SRC's floor: the build image's
# tacplus refuses it otherwise. The justfile's gates run against SRC's source, and its release tests from
# its own directory (justfile, release-test), since SRC may have no ci/.
#
# Usage:  ci/release-file.sh SRC FILE     prints the path of the FILE to use
set -euo pipefail

src="$(cd "$1" && pwd)"
file="$2"
if [[ -f "$src/$file" ]]; then
  echo "$src/$file"
  exit 0
fi
own="$(cd "$(dirname "$0")/.." && pwd)/$file"
[[ -f "$own" ]] || { echo "release-file.sh: neither $src nor the workflow's commit has a $file" >&2; exit 1; }
echo "release-file.sh: $src has no $file: using the workflow's commit's, $own" >&2
echo "$own"
