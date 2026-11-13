#!/usr/bin/env bash
# Tests of the release scripts (.github/workflows/release.yml) that need no container: the image tags
# (ci/release-tags.sh), the notes (ci/release-notes.sh) and the tacplus.lock a release builds with
# (ci/release-lock.sh), each in a scratch repository with the tags a case needs.
#
# Usage:  ci/release_test.sh
set -euo pipefail

ci="$(cd "$(dirname "$0")" && pwd)"
scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT
failed=0

check() { # name got want
  if [[ "$2" == "$3" ]]; then
    echo "ok   $1"
  else
    echo "FAIL $1: got [$2], want [$3]"
    failed=1
  fi
}

# A repository whose commits are named by their subjects, each tagged as given: "subject[:tag]"...
repo() {
  local dir="$scratch/$1"
  shift
  git init -q "$dir"
  git -C "$dir" config user.email test@example.com
  git -C "$dir" config user.name test
  for c in "$@"; do
    git -C "$dir" commit -q --allow-empty -m "${c%%:*}"
    [[ "$c" == *:* ]] && git -C "$dir" tag -a -m "${c#*:}" "${c#*:}"
  done
  echo "$dir"
}

tags() { (cd "$1" && "$ci/release-tags.sh" "$2"); }
notes() { (cd "$1" && "$ci/release-notes.sh" "$2"); }

# The first release: its version and the line tag.
r="$(repo first "one" "two:v0.1.0")"
check "first release: X.Y.Z and X.Y" "$(tags "$r" v0.1.0)" "0.1.0 0.1"
check "first release: notes list every commit" "$(notes "$r" v0.1.0 | grep -c '^- ')" "2"
check "first release: notes say Changes" "$(notes "$r" v0.1.0 | head -1)" "Changes:"

# A patch on the highest line moves that line's tag and no other.
r="$(repo patch "one:v0.1.0" "fix:v0.1.1")"
check "v0.1.1 after v0.1.0: moves 0.1" "$(tags "$r" v0.1.1)" "0.1.1 0.1"
check "v0.1.0 after v0.1.1: does not move 0.1 back" "$(tags "$r" v0.1.0)" "0.1.0"
check "v0.1.1 notes: since v0.1.0" "$(notes "$r" v0.1.1)" "$(printf 'Changes since v0.1.0:\n\n- fix (%s)' "$(git -C "$r" rev-parse --short v0.1.1^{commit})")"

# A patch on an older line moves that line, not the newer one; the newer one's release does not move the older.
r="$(repo lines "one:v0.1.0" "feature:v0.2.0" "backport:v0.1.1")"
check "v0.1.1 beside v0.2.0: moves 0.1 only" "$(tags "$r" v0.1.1)" "0.1.1 0.1"
check "v0.2.0 beside v0.1.1: moves 0.2 only" "$(tags "$r" v0.2.0)" "0.2.0 0.2"
check "v0.1.1 notes: since v0.1.0, not v0.2.0" "$(notes "$r" v0.1.1 | head -1)" "Changes since v0.1.0:"

# By version, not lexically: v0.1.10 is above v0.1.9.
r="$(repo numeric "a:v0.1.9" "b:v0.1.10")"
check "v0.1.10 above v0.1.9: moves 0.1" "$(tags "$r" v0.1.10)" "0.1.10 0.1"
check "v0.1.9 below v0.1.10: does not" "$(tags "$r" v0.1.9)" "0.1.9"
check "v0.1.10 notes: since v0.1.9" "$(notes "$r" v0.1.10 | head -1)" "Changes since v0.1.9:"

# A pre-release gets only its own tag, and is not the previous release of the next one.
r="$(repo pre "a:v0.1.0" "b:v0.2.0-rc.1" "c:v0.2.0")"
check "pre-release: its own tag only" "$(tags "$r" v0.2.0-rc.1)" "0.2.0-rc.1"
check "after a pre-release: since the last release" "$(notes "$r" v0.2.0 | head -1)" "Changes since v0.1.0:"

# Not a version.
r="$(repo bad "a:v1")"
check "release-tags.sh refuses v1" "$(tags "$r" v1 2>/dev/null || echo refused)" "refused"
check "release-notes.sh refuses v1" "$(notes "$r" v1 2>/dev/null || echo refused)" "refused"

# The lock: the source's own, else the one beside these scripts (a tag cut before tacplus.lock existed).
mkdir -p "$scratch/locked" "$scratch/unlocked"
touch "$scratch/locked/tacplus.lock"
check "lock: the source's own" "$("$ci/release-lock.sh" "$scratch/locked")" "$scratch/locked/tacplus.lock"
check "lock: none in the source, the workflow's" "$("$ci/release-lock.sh" "$scratch/unlocked" 2>/dev/null)" "$(cd "$ci/.." && pwd)/tacplus.lock"
check "lock: none in the source, says so" "$("$ci/release-lock.sh" "$scratch/unlocked" 2>&1 >/dev/null | grep -c 'has no tacplus.lock')" "1"

exit "$failed"
