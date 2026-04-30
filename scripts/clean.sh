#!/usr/bin/env bash

# This script deletes unnecessary files.
# Edit the arrays to customize it.

set -euo pipefail

# Common directories to delete
/bin/rm -rfv tmp-*
/bin/rm -rfv dist/
/bin/rm -rfv .output/
/bin/rm -rfv .tanstack/

# node_modules cache
/bin/rm -rfv ./node_modules/.cache/prettier/.prettier-cache

# Directories to exclude from traversal
excluded_dirs=(
  '.claude'
  '.worktrees'
)

# Files deleted by find
find_files=(
  '.DS_Store'
  'Thumbs.db'
  '*.pyc'
  '*.pyo'
  '*~'
)

# Directories deleted by find
find_dirs=(
  '__pycache__'
)

# ======= You don't need to edit below this line =======

build_name_expr() {
  local -n arr=$1
  local out=()

  for pattern in "${arr[@]}"; do
    out+=( -name "$pattern" -o )
  done

  unset 'out[-1]'
  printf '%q ' "${out[@]}"
}

build_excluded_expr() {
  local out=()

  for dir in "${excluded_dirs[@]}"; do
    out+=( -path "./$dir" -o -path "./$dir/*" -o )
  done

  unset 'out[-1]'
  printf '%q ' "${out[@]}"
}

excluded_expr=$(build_excluded_expr)
file_expr=$(build_name_expr find_files)
dir_expr=$(build_name_expr find_dirs)

eval "find . \
  \( $excluded_expr \) -prune -o \
  \( -type f \( $file_expr \) -print -exec /bin/rm -fv {} + \) -o \
  \( -type d \( $dir_expr \) -print -exec /bin/rm -rfv {} + \)"
