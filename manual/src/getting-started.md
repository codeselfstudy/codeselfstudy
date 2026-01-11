# Getting Started

## Prerequisites

- Install [Bun](https://bun.sh/). It's an alternative to Node.js.

## Install and Run

```bash
# Install dependencies
bun install

# Run the dev server
bun run dev
```

The dev server runs at `http://localhost:7001`.

## Build and Preview

```bash
bun run build
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
bun run db:generate
bun run db:migrate
bun run db:studio
```
