#!/usr/bin/env bash
# The gate of the ping page (web/, README): built as the fleet builder's webapp recipe builds it (tacplus
# docs/USER.md, "Contributing to the gateway"), from the working directory's source. The root tacplus.yaml
# contributes web/ as its webapp, with the recipe's inputs; its client generates with the compiler, leaving
# every tracked file as it found it; `npm ci --ignore-scripts`; its tests; and `npm run build` with
# TACPLUS_WEB_BASE=/-/apps/tenantecho/, twice from a clean node_modules, to the same bytes in web/dist/.
#
# Usage:  ci/web_test.sh        (TACPLUS_BIN names the compiler, default tacplus)
set -euo pipefail

tacplus="${TACPLUS_BIN:-tacplus}"
base=/-/apps/tenantecho/
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

[[ -f tacplus.yaml && -d web ]] || { echo "web_test.sh: run it at the repository's root" >&2; exit 1; }

# web/ is a project of its own, so it has its own lock: the root's pin, moved with it (`tacplus pin` in both).
check "web: tacplus.lock is the root's" "$(cmp -s tacplus.lock web/tacplus.lock && echo same)" "same"

before="$(git diff HEAD | sha256sum)"
(cd web && "$tacplus" build)
check "web: generating the client leaves tracked files as they were" "$(git diff HEAD | sha256sum)" "$before"

build() { # into DIR: a clean install and the recipe's build
  rm -rf web/node_modules web/dist
  (cd web && npm ci --ignore-scripts --no-audit --no-fund)
  (cd web && TACPLUS_WEB_BASE="$base" npm run build)
  (cd web/dist && find . -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum) > "$1"
}

build "$scratch/first"
(cd web && npx --no-install vitest run)

# The contract, read as the fleet build reads it: js-yaml is the page's own dependency, just installed.
contract() {
  (cd web && node --input-type=module -e '
    import fs from "node:fs"
    import yaml from "js-yaml"
    const y = yaml.load(fs.readFileSync("../tacplus.yaml", "utf8"))
    const w = (y.contribute || {}).webapp || {}
    const p = JSON.parse(fs.readFileSync(w.dir ? `../${w.dir}/package.json` : "package.json", "utf8"))
    console.log([y.kind, w.dir, w.title, Object.keys(p.scripts || {}).join(",")].join(" "))
  ')
}
check "tacplus.yaml: a service contributing web/ as Echo, whose only npm script is build" "$(contract)" "service web Echo build"
check "web: the recipe's inputs are committed" "$(git ls-files web/package.json web/package-lock.json | wc -l)" "2"
check "web: dist/ is not committed" "$(git ls-files web/dist | wc -l)" "0"
check "web: the build has an index.html under its base" "$(grep -c "src=\"$base" web/dist/index.html)" "1"

build "$scratch/second"
check "web: two builds write the same bytes" "$(diff "$scratch/first" "$scratch/second" && echo same)" "same"
echo "web: $(wc -l < "$scratch/first") files in web/dist/"

exit "$failed"
