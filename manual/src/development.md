# Development

Notes for working on the codebase day-to-day.

## Running the app

```bash
just dev
```

Boots two servers in parallel:

- `apps/web/` (TanStack Start dev server) on `http://localhost:7001`.
- `apps/api/` (Go + Echo) on `http://localhost:8080`.

The Vite dev server proxies `/api` and `/ws` to `:8080`, so you hit the same paths in dev as in production.

To run only one side:

```bash
just dev-web   # Vite only
just dev-api   # Go only — sources .env.local so /api/me and /api/todos mount
```

## Updating dependencies

```bash
# JS dependencies (apps/web)
bun update --interactive

# Go dependencies (apps/api)
cd apps/api && go get -u ./... && go mod tidy
```

## Running tests

```bash
just test            # Go race tests + Vitest
just test-watch      # Vitest in watch mode (web only)
just test-coverage   # Vitest with coverage report
```

Backend changes require Go test coverage — that's a hard rule of the stack. Add tests under `apps/api/` next to the code under test (`*_test.go`).

## Linting and formatting

```bash
just format    # Prettier across the repo
just lint      # ESLint + Stylelint on the web side
```

`lefthook` runs Prettier on staged files in `pre-commit` and `just test` + lint in `pre-push`.

## See available recipes

```bash
just
```
