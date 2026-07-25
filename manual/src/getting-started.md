# Getting Started

## Prerequisites

- Install [Bun](https://bun.com/). It's an alternative to Node.js.
- Install [Go](https://go.dev/) 1.26+. The backend is Go + Echo.
- Install [just](https://just.systems/). It's a command runner.

## Install and Run

```bash
# Install dependencies
bun install

# Run the web (:7001) and Go API (:8080) dev servers together.
# They run independently (no dev proxy): the navbar's same-origin /api/me
# hits Astro on :7001 (which has no API); the authed API needs the Go build.
just dev
```

The dev server runs at `http://localhost:7001`.

## Build and Preview

```bash
just build
bun run --filter web preview
```

`astro preview` serves only the prerendered files. To preview production URL behavior too (redirects, trailing-slash canonicalization, `/api/*`), run the Go server against the same build instead: `just build`, then `just dev-api`.

## Environment Variables

Copy `.env.local.example` to `.env.local` at the repo root and fill in the values. The example file is the authoritative, commented reference; the short version:

```bash
# database (also enables user accounts when set)
DATABASE_URL="dev.db"

# WorkOS auth — all server-only; sign-in is a server-side session, so the
# browser bundle reads no WorkOS values and there are no PUBLIC_/VITE_ aliases.
# All five must be set or the /auth/* routes stay off.
WORKOS_CLIENT_ID=client_...
WORKOS_API_HOSTNAME=api.workos.com
WORKOS_API_KEY=sk_test_...
WORKOS_COOKIE_PASSWORD=  # 32+ secure random chars, e.g. from: openssl rand -base64 32
APP_BASE_URL=http://localhost:8080

# optional: Slack webhook pinged when a user requests account deletion
SLACK_WEBHOOK_FOR_ADMIN_CHANNEL=

# email-ingest pipeline (optional — see below)
INGEST_TOKEN=""
GEMINI_API_KEY=""
SLACK_WEBHOOK_FOR_DEALS_CHANNEL=""
```

`DATABASE_URL` is used by the migrate CLI in `apps/api/cmd/migrate`, by tests, and — when the email-ingest pipeline is enabled — by the runtime server. In production it is a Turso URL (`libsql://<db>.turso.io?authToken=<token>`). The auth token may instead be supplied separately as `TURSO_AUTH_TOKEN` (Turso's own convention), in which case `DATABASE_URL` can be a bare `libsql://<db>.turso.io`.

The email-ingest pipeline (the `apps/email_receiver` Worker → `/api/ingest` → Turso → Slack) is opt-in: it runs only when `DATABASE_URL` and `INGEST_TOKEN` are both set, and extraction additionally needs `GEMINI_API_KEY`. See [Architecture](./architecture.md#email--deals--slack) for the full flow.

## Database Tasks

The Go side owns the schema. Migrations live under `apps/api/internal/db/migrations/`. A local database is migrated automatically on server startup, so most of the time you don't need to run anything; in production a remote Turso database is migrated by the deploy's release command instead (see [Deployment](./deployment.md#migrations)).

```bash
just db_status               # show which migrations have been applied
just db_migrate              # apply pending migrations now
just db_down                 # roll back the most recent migration (dev only)
just db_create add_users     # scaffold a new migration file
```

These are thin wrappers around the embedded goose CLI in
`apps/api/cmd/migrate`.
