# Getting Started

## Prerequisites

- Install [Bun](https://bun.sh/). It's an alternative to Node.js.
- Install [just](https://just.systems/). It's a command runner.

## Install and Run

```bash
# Install dependencies
bun install

# Run the dev server
just dev
```

The dev server runs at `http://localhost:7001`.

## Build and Preview

```bash
just build
bun run preview
```

## Environment Variables

Environment variables are validated in `src/env.ts`. Create a `.env.local` file:

```bash
# client-side variables
VITE_WORKOS_CLIENT_ID="your_workos_client_id"
VITE_WORKOS_API_HOSTNAME="your_workos_api_hostname"

# server-only variables
SERVER_URL="http://localhost:7001"
WORKOS_API_KEY="your_workos_api_key"
DATABASE_URL="dev.db"
TURSO_AUTH_TOKEN="your_turso_auth_token"
```

`DATABASE_URL` is used by Drizzle with SQLite.

To consume validated variables in code:

```ts
import { env } from "@/env";

console.log(env.SERVER_URL);
```

## Database Tasks

```bash
just db_generate
just db_migrate
just db_studio
just db_push
just db_pull
```

These are wrappers around `drizzle-kit` that ensure it runs under Bun.
