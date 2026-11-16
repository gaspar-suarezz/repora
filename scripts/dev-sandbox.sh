#!/usr/bin/env bash
#
# Build a disposable Repora sandbox: four repositories in different states, a
# local bare remote, and a configuration that points at them.
#
# Usage: scripts/dev-sandbox.sh [root]   (default: /tmp/repora-sandbox)
#
# Safe: everything is created under the given root and removed first.

set -euo pipefail

ROOT="${1:-/tmp/repora-sandbox}"
CONFIG="$ROOT/config.yaml"

rm -rf "$ROOT"
mkdir -p "$ROOT/code" "$ROOT/remotes"

make_repo() {
	local name="$1"
	local dir="$ROOT/code/$name"
	mkdir -p "$dir"
	git -C "$dir" init -q -b main
	git -C "$dir" config user.name "Repora Sandbox"
	git -C "$dir" config user.email "sandbox@localhost"
	printf '# %s\n' "$name" >"$dir/README.md"
	git -C "$dir" add .
	git -C "$dir" commit -qm "initial commit"
}

make_repo api
make_repo web
make_repo infra
make_repo shared

# api: one staged file, one unstaged file, one untracked file.
printf 'api\nchanged\n' >"$ROOT/code/api/README.md"
printf 'staged content\n' >"$ROOT/code/api/staged.txt"
git -C "$ROOT/code/api" add staged.txt
printf 'untracked\n' >"$ROOT/code/api/notes.txt"

# web: clean.

# infra: local bare remote with the branch pushed, then a local commit (ahead).
git init -q --bare "$ROOT/remotes/infra.git"
git -C "$ROOT/code/infra" remote add origin "$ROOT/remotes/infra.git"
git -C "$ROOT/code/infra" push -q -u origin main
printf 'change\n' >"$ROOT/code/infra/change.txt"
git -C "$ROOT/code/infra" add .
git -C "$ROOT/code/infra" commit -qm "local change"

# shared: merge conflict on main.
printf 'base\n' >"$ROOT/code/shared/conflict.txt"
git -C "$ROOT/code/shared" add .
git -C "$ROOT/code/shared" commit -qm "base"
git -C "$ROOT/code/shared" checkout -qb other
printf 'other\n' >"$ROOT/code/shared/conflict.txt"
git -C "$ROOT/code/shared" commit -qam "other"
git -C "$ROOT/code/shared" checkout -q main
printf 'main\n' >"$ROOT/code/shared/conflict.txt"
git -C "$ROOT/code/shared" commit -qam "main"
git -C "$ROOT/code/shared" merge other >/dev/null 2>&1 || true

cat >"$CONFIG" <<YAML
version: 1

repositories:
  - { id: api, name: API, path: $ROOT/code/api }
  - { id: web, name: Web, path: $ROOT/code/web }
  - { id: infra, name: Infrastructure, path: $ROOT/code/infra }
  - { id: shared, name: Shared, path: $ROOT/code/shared }

workspaces:
  - id: sandbox
    name: Sandbox
    description: Disposable repositories with different states
    repositories: [api, web, infra, shared]
  - id: empty
    name: Empty
    description: A workspace with no repositories
    repositories: []

preferences:
  theme: dark
  refresh_interval: 30
  status_workers: 4
YAML

printf '\nSandbox ready: %s\n' "$ROOT"
printf 'Config:        %s\n\n' "$CONFIG"
printf '  ./bin/repora --config %s\n' "$CONFIG"
printf '  ./bin/repora --config %s --debug-status sandbox\n\n' "$CONFIG"
