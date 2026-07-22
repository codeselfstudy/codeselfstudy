# Getting Started

## Prerequisites

- Install [Bun](https://bun.sh/). It's an alternative to Node.js.
- Install [Go](https://go.dev/) 1.26+. The backend is Go + Echo.
- Install [just](https://just.systems/). It's a command runner.

## Install and Run

```bash
# Install dependencies
bun install

# Run the web (:7001) and Go API (:8080) dev servers together.
# They run independently (no dev proxy); the navbar's /api/me call only
# resolves against the Go server, not the standalone Astro server.
just dev
```

The dev server runs at `http://localhost:7001`.

## Build and Preview

```bash
just build
bun run preview
```

## Environment Variables

Copy `env.local.example` to `.env.local` at the repo root and fill in:

```bash
# client-side variables (browser-exposed). The Astro app consumes these for the
# WorkOS AuthKit sign-in control; the Go verifier reads them as fallback names.
VITE_WORKOS_CLIENT_ID="your_workos_client_id"
VITE_WORKOS_API_HOSTNAME="your_workos_api_hostname"

# server-only variables
WORKOS_API_KEY="your_workos_api_key"
DATABASE_URL="dev.db"

# email-ingest pipeline (optional — see below)
INGEST_TOKEN=""
GEMINI_API_KEY=""
SLACK_WEBHOOK_FOR_DEALS_CHANNEL=""
```

`DATABASE_URL` is used by the migrate CLI in `apps/api/cmd/migrate`, by tests, and — when the email-ingest pipeline is enabled — by the runtime server. In production it is a Turso URL (`libsql://<db>.turso.io?authToken=<token>`). The auth token may instead be supplied separately as `TURSO_AUTH_TOKEN` (Turso's own convention), in which case `DATABASE_URL` can be a bare `libsql://<db>.turso.io`.

The email-ingest pipeline (the `apps/email_receiver` Worker → `/api/ingest` → Turso → Slack) is opt-in: it runs only when `DATABASE_URL` and `INGEST_TOKEN` are both set, and extraction additionally needs `GEMINI_API_KEY`. See [Architecture](./architecture.md#email--deals--slack) for the full flow.

## Database Tasks

The Go side owns the schema. Migrations live under
`apps/api/internal/db/migrations/` and are applied on server startup, so most
of the time you don't need to run anything.

```bash
just db_status               # show which migrations have been applied
just db_migrate              # apply pending migrations now
just db_down                 # roll back the most recent migration (dev only)
just db_create add_users     # scaffold a new migration file
```

These are thin wrappers around the embedded goose CLI in
`apps/api/cmd/migrate`.
