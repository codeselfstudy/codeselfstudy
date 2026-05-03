#!/usr/bin/env bash

# Deletes build artifacts and stray junk files. Edit the lists at the top to
# customize. Kept simple so it runs under bash 3.2 (macOS default).

set -euo pipefail

# Web build outputs (TanStack Start + Vite + Nitro).
/bin/rm -rfv apps/web/.output
/bin/rm -rfv apps/web/.tanstack
/bin/rm -rfv apps/web/dist

# Go static-asset mirror (populated by `just build` from .output/public).
/bin/rm -rfv apps/api/static

# Local Go binaries from `go build .` if anyone runs them.
/bin/rm -fv apps/api/api apps/api/server

# Stray top-level dirs that older configs sometimes left behind.
/bin/rm -rfv tmp-*

# Prettier cache inside hoisted node_modules.
/bin/rm -rfv ./node_modules/.cache/prettier/.prettier-cache

# OS / editor junk anywhere in the tree, except inside the listed dirs.
find . \
  \( -path './.claude' -o -path './.claude/*' \
  -o -path './.worktrees' -o -path './.worktrees/*' \
  -o -path './node_modules' -o -path './node_modules/*' \
  -o -path './apps/*/node_modules' -o -path './apps/*/node_modules/*' \
  -o -path './.git' -o -path './.git/*' \) -prune -o \
  \( -type f \( -name '.DS_Store' -o -name 'Thumbs.db' \
  -o -name '*.pyc' -o -name '*.pyo' -o -name '*~' \) \
  -print -exec /bin/rm -fv {} + \) -o \
  \( -type d -name '__pycache__' -print -exec /bin/rm -rfv {} + \)
