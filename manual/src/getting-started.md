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
# The two run independently; the static site doesn't call the API yet.
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
# client-side variables (browser-exposed). The Astro app doesn't consume
# these yet; the Go verifier reads them as fallback names.
VITE_WORKOS_CLIENT_ID="your_workos_client_id"
VITE_WORKOS_API_HOSTNAME="your_workos_api_hostname"

# server-only variables
WORKOS_API_KEY="your_workos_api_key"
DATABASE_URL="dev.db"
```

`DATABASE_URL` is used by the migrate CLI in `apps/api/cmd/migrate` and by
tests. Once the first DB-backed endpoint lands it will also be read at
server startup.

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
