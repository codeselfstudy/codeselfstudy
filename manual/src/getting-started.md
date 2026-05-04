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
# Vite proxies /api and /ws to the Go server.
just dev
```

The dev server runs at `http://localhost:7001`.

## Build and Preview

```bash
just build
bun run preview
```

## Environment Variables

Web-side environment variables are validated in `apps/web/src/env.ts`. Copy
`env.local.example` to `.env.local` at the repo root and fill in:

```bash
# client-side variables (browser-exposed)
VITE_WORKOS_CLIENT_ID="your_workos_client_id"
VITE_WORKOS_API_HOSTNAME="your_workos_api_hostname"

# server-only variables
SERVER_URL="http://localhost:7001"
WORKOS_API_KEY="your_workos_api_key"
DATABASE_URL="dev.db"
```

`DATABASE_URL` is read by the Go API and points to a local SQLite file (or
`:memory:` in tests).

To consume validated variables in TypeScript code:

```ts
import { env } from "@/env";

console.log(env.SERVER_URL);
```

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
