# Deployment

The deployed app is a single Go binary that serves the prerendered web app, the JSON API at `/api/*`, and (in future) a WebSocket endpoint at `/ws`. There is no JavaScript runtime in production — the bundle is built locally and shipped as static files alongside the binary.

## Architecture in production

- `apps/web/dist/` — prerendered HTML/CSS/JS produced by `bun run build`. Built locally so Fly's remote builder doesn't re-run `bun install` (~10 min on a cold cache).
- `apps/api/static/` — copy of `apps/web/dist/` mirrored by `just build` so the Go binary serves it.
- `server` (the Go binary) — Echo + middleware + auth + DB; embeds the goose migrations and applies them via the deploy's release command (see [Migrations](#migrations)).

## Docker image

`Dockerfile` is multi-stage:

1. **Build stage** (`golang:1.26-alpine`): copy `apps/api/`, copy the prebuilt `apps/web/dist` into `./static`, then `go build -trimpath -ldflags="-s -w" -o /out/server .` with `CGO_ENABLED=0`.
2. **Runtime stage** (`gcr.io/distroless/static-debian12:debug-nonroot`): copy `/server` and `/static`. Distroless gives CA certs, a nonroot user, and a BusyBox shell for `fly ssh console` — about 2 MB on disk and zero steady-state RAM beyond the binary itself.

The image runs as `nonroot`, exposes `8080`, and `ENTRYPOINT ["/server"]`.

## Fly.io

- `fly.toml` targets a 256 MB shared-cpu-1x machine with auto-stop enabled.
- `/healthz` returns 204 — used by Fly's health checks.

To deploy:

```bash
just deploy
```

`just deploy` runs `just build` first (so the prerendered site is up to date), then `fly deploy`. Fly's remote builder pulls the prebuilt `apps/web/dist` from the build context — no JS install on the remote side.

## Migrations

Goose migrations live in `apps/api/internal/db/migrations/` and are embedded into the binary via `//go:embed`. In production they are applied by a dedicated deploy step, not on server startup: `fly.toml` sets `[deploy] release_command = "/server -migrate"`, so Fly runs `server -migrate` once per deploy in a temporary machine before the new version serves traffic. If it fails, the release command fails — which aborts the deploy and leaves the previous version serving, rather than crash-looping the new one. Migrations are idempotent (goose tracks state in its own table), so the step is a no-op when the schema is already current.

The one exception is a local SQLite database, which the server migrates on boot as a dev convenience; a remote Turso database is always migrated out of band by the release command.

For ad-hoc runs against a local DB:

```bash
just db_status
just db_migrate
just db_create add_some_table   # scaffolds a new .sql file
```

## Forcing a Slack digest

The deals pipeline posts to Slack at most once per `DIGEST_INTERVAL` (default 24h), and a digest is only attempted when an email is ingested — there is no scheduler. So deals from a newsletter that lands inside the interval sit queued (`posted_in_digest_id IS NULL`) until the next email is forwarded after the window opens. To flush the queue immediately — or to recover a digest whose Slack post failed — force one:

```bash
curl -sS -X POST https://codeselfstudy.com/api/admin/digest \
  -H "Authorization: Bearer $INGEST_TOKEN"
```

`POST /api/admin/digest` runs the same digest as an ingest but skips the interval check; the stale-claim guard still holds, so concurrent calls can't double-post. `INGEST_TOKEN` is the Fly secret shared with the Worker — the same value is in `.env.local`, so for a local run point the URL at `http://localhost:8080` instead.

Responses:

- `{"posted":true}` — a digest with the queued deals was posted to `SLACK_WEBHOOK_FOR_DEALS_CHANNEL`.
- `{"posted":false}` — nothing to post (no queued deals, or a digest is mid-flight). Safe to repeat.
- `500 {"message":"digest failed"}` — the Slack post itself failed; the deals stay queued for a later attempt. Check `just logs`.

To confirm what happened, query the deployed database read-only:

```bash
turso db shell codeselfstudy "SELECT count(*) FROM deals WHERE posted_in_digest_id IS NULL;"
turso db shell codeselfstudy "SELECT id, posted_at, status, deal_count FROM digests ORDER BY id DESC LIMIT 1;"
```

A successful force-post drops the unposted count to 0 and adds a `status=posted` digest row.
