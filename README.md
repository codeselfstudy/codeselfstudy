# Code Self Study Website

This is the new [Code Self Study](https://codeselfstudy.com/) website.

Attend a meetup to find out how to contribute. :construction:

See the documentation in [manual](./manual/src/SUMMARY.md). If [just](https://just.systems/) is installed, you can view the manual in the browser with the command: `just manual`.

## Architecture

Bun workspace with two apps and a Go binary that fronts the whole thing in production.

```
codeselfstudy/
├── apps/
│   ├── web/                  TanStack Start (Vite + React + TS).
│   │   └── .output/public/   Prerendered static site (build-time output).
│   └── api/                  Go + Echo backend.
│       └── static/           Mirrored from web's .output/public at build time.
├── Dockerfile                Multi-stage: Go build + distroless runtime.
├── fly.toml                  256 MB shared-cpu-1x.
└── justfile                  dev / build / test / deploy.
```

The Go binary serves the prerendered HTML, the JSON API (`/api/*`), and a future WebSocket endpoint at `/ws`. No JS runtime in production. Targets a single 256 MB Fly machine.

**Auth.** WorkOS AuthKit on the client; an Echo middleware validates access tokens against the WorkOS JWKS for protected `/api/*` routes.

**Database.** SQLite via `modernc.org/sqlite` (pure Go, no CGO). Drizzle owns the schema in `apps/web/src/db/schema.ts` and runs migrations from the web side; the Go API reads and writes through plain SQL. Remote Turso (libsql://) is a follow-up.

**Build.** `bun run build` prerenders all routes; `just build` mirrors the output into `apps/api/static/` so the Go binary picks it up. The Docker build runs locally and ships the prebuilt artifact in the build context — Fly's remote builder doesn't re-run `bun install`.

## Quick start

Requires [Bun](https://bun.com/), [Go](https://go.dev/) 1.26+, and [just](https://just.systems/).

```sh
# clone, then:
bun install
cp env.local.example .env.local   # fill in WorkOS + DATABASE_URL
just dev                          # web :7001, api :8080 (Vite proxies /api and /ws)
```

```sh
# Common tasks
just build       # produce a deployable artifact
just test        # Go race tests + Vitest
just deploy      # build, then fly deploy
```

See [AGENTS.md](AGENTS.md) for the full developer guide.

## Goals:

- help people in the group find something in common to work on
- meetup activity

### Ideas so far:

- send a coding puzzle of the day into the forum, slack, and/or the browser extension so that interested people have a common task to discuss
- add new puzzles (links to other sites or original puzzles)
- mark puzzles that you've completed
- saving puzzles to do later
- use the browser extension to add new puzzles to the database?
- fetch puzzle by difficulty and type of problem
- commenting? (forum integration or separate)
- voting
- browser extension integration

## Contributing

Contributions are welcome! This project is developed openly on GitHub, and we are happy to accept help in many forms — including frontend changes, bug fixes, new ideas, or improving documentation.

For more details, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Licenses

The code is licensed under BSD 3-Clause license. Some of the subdirectories are not licensed under BSD 3-Clause license. To see the licenses of individual subdirectories, please look for the `LICENSE.md` files in subdirectories.

You can use this command to see their locations:

```bash
find . -name "LICENSE.md" -not -path "*/node_modules/*"
```

or with `just`:

```bash
just find_licenses
```

Basically, all the computer code is BSD 3-Clause licensed, except that the website's rendered text content and images are not licensed and cannot be reused, because they are unique to this website and brand (Code Self Study).
