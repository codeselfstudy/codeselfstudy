#!/usr/bin/env bash

# End-to-end smoke test for the built site served by the Go server.
#
# Assumes `just build` has already populated apps/api/static/ (the `just
# smoke_test` recipe runs the build first). Builds the Go binary, starts it on a
# test port, curls the served site, and tears the server down. Exits non-zero if
# any check fails. Kept simple so it runs under bash 3.2 (macOS default).

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PORT="${SMOKE_PORT:-8091}"
BASE="http://localhost:${PORT}"

if [ ! -f "$ROOT/apps/api/static/index.html" ]; then
  echo "apps/api/static/index.html not found — run 'just build' first." >&2
  exit 1
fi

BIN="$(mktemp -t smoke-server.XXXXXX)"
LOG="$(mktemp -t smoke-server-log.XXXXXX)"
( cd "$ROOT/apps/api" && go build -o "$BIN" . )
( cd "$ROOT/apps/api" && PORT="$PORT" exec "$BIN" ) >"$LOG" 2>&1 &
SERVER_PID=$!

cleanup() {
  kill "$SERVER_PID" 2>/dev/null || true
  wait "$SERVER_PID" 2>/dev/null || true
  /bin/rm -f "$BIN" "$LOG"
}
trap cleanup EXIT

# Wait for the server to answer /healthz.
ready=0
for _ in $(seq 1 40); do
  if curl -fsS -o /dev/null "${BASE}/healthz" 2>/dev/null; then
    ready=1
    break
  fi
  sleep 0.25
done
if [ "$ready" -ne 1 ]; then
  echo "server did not become ready on ${BASE}; log:" >&2
  cat "$LOG" >&2
  exit 1
fi

FAILURES=0

check_status() { # path want-status
  local got
  got="$(curl -s -o /dev/null -w "%{http_code}" "${BASE}$1")"
  if [ "$got" = "$2" ]; then
    echo "  ok    $1 -> $got"
  else
    echo "  FAIL  $1 -> $got (want $2)"
    FAILURES=$((FAILURES + 1))
  fi
}

check_redirect() { # path want-status want-location-path
  local out code loc path
  out="$(curl -s -o /dev/null -w "%{http_code} %{redirect_url}" "${BASE}$1")"
  code="${out%% *}"
  loc="${out#* }"
  # Strip scheme://host to compare the Location path (+ query) exactly, so a
  # redirect to the wrong page can't slip through a loose suffix match.
  path="/${loc#*://*/}"
  if [ "$code" = "$2" ] && [ "$path" = "$3" ]; then
    echo "  ok    $1 -> $code $path"
  else
    echo "  FAIL  $1 -> $code $path (want $2 $3)"
    FAILURES=$((FAILURES + 1))
  fi
}

check_body_contains() { # path needle
  if curl -s "${BASE}$1" | grep -q "$2"; then
    echo "  ok    $1 contains '$2'"
  else
    echo "  FAIL  $1 missing '$2'"
    FAILURES=$((FAILURES + 1))
  fi
}

echo "Pages (200 at trailing-slash URLs):"
for p in / /about/ /codewars/ /contact/ /credits/ /discounts/ /events/ \
  /forum/ /jobs/ /learn/ /puzzles/ /s/ /settings/ /tools/; do
  check_status "$p" 200
done

echo "Trailing-slash canonicalization:"
check_redirect /about 301 /about/

echo "Legacy redirects:"
check_redirect /book 308 /learn/          # exact
check_redirect /blog/anything 308 /learn/ # wildcard
check_redirect /index.html 308 /          # precedence over the real file

echo "Other:"
check_body_contains /s/ noindex
check_body_contains /settings/ noindex
check_status /nonexistent-page-xyz/ 404
check_body_contains /nonexistent-page-xyz/ "Page not found"
check_status /healthz 204
check_status /sitemap-index.xml 200

echo
if [ "$FAILURES" -ne 0 ]; then
  echo "SMOKE FAILED: $FAILURES check(s) failed."
  exit 1
fi
echo "SMOKE OK"
