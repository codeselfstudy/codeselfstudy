# Tooling & Quality

## Scripts

Most day-to-day work goes through `just` — see [Scripts](./scripts.md). The
underlying commands:

```bash
# Web (apps/web)
bun run --filter web test            # Vitest
bun run --filter web test:watch      # Vitest in watch mode
bun run --filter web test:coverage   # Vitest with coverage report
bun run --filter web lint            # ESLint + Stylelint

# API (apps/api)
go test -race ./...                  # Go race tests
go vet ./...                         # static analysis

# Repo-wide
bun run format                       # Prettier
bun run check                        # Prettier + ESLint
```

## Linting and Formatting

- ESLint uses a flat config (`typescript-eslint`, `eslint-plugin-astro`, `react-hooks`) on the web side.
- Stylelint checks CSS and Tailwind usage.
- Prettier is the standard formatter.
- Styling is done with Tailwind CSS.
- Go follows `gofmt` and `go vet`.

## Tests and Coverage

- **Web:** Vitest (`apps/web/vitest.config.ts`). `just test-coverage` produces an HTML report under `apps/web/coverage/`. `apps/web/src/lib/**` enforces 100% line/branch/function/statement coverage.
- **API:** stdlib `testing` + `httptest`. All Go code requires test coverage — this is a hard rule of the stack experiment. Tests live next to the code as `*_test.go`.

## Continuous Integration

`.github/workflows/test.yml` runs both sides on every pull request and on pushes to `main`:

- `bun install` + `bun run --filter web test`
- `cd apps/api && go test -race ./... && go vet ./...`
- ESLint and Stylelint via `bunx` against `apps/web`.

## Git Hooks

`lefthook.yml` runs Prettier on staged files in `pre-commit` and `just test` + lint in `pre-push`. If hooks block you, fix the reported files and re-run the command — never bypass them with `--no-verify`.
