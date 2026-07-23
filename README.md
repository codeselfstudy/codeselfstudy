# Code Self Study Website

This is the new [Code Self Study](https://codeselfstudy.com/) website.

Attend a meetup to find out how to contribute. :construction:

See the documentation in [manual](./manual/src/SUMMARY.md). If [just](https://just.systems/) is installed, you can view the manual in the browser with the command: `just manual`.

## Architecture

Bun workspace with three apps: an Astro frontend, a Go + Echo backend that fronts the whole thing in production, and a Cloudflare Worker that receives forwarded email.

```
codeselfstudy/
├── apps/
│   ├── web/                  Astro (SSG, React islands).
│   │   └── dist/             Prerendered static site (build-time output).
│   ├── api/                  Go + Echo backend (serves the site + JSON API).
│   │   └── static/           Mirrored from web's dist/ at build time.
│   └── email_receiver/       Cloudflare Worker: forwards "deals" email to /api/ingest.
├── Dockerfile                Multi-stage: Go build + distroless runtime.
├── fly.toml                  256 MB shared-cpu-1x.
└── justfile                  dev / build / test / deploy.
```

The Go binary serves the prerendered HTML, the JSON API (`/api/*`), and a future WebSocket endpoint at `/ws`. It also owns the URL layer: legacy redirects and trailing-slash canonicalization (the site is built with `trailingSlash: "always"`) live in the Go server, since a static build can't emit real redirects. No JS runtime in production. Targets a single 256 MB Fly machine.

**Auth.** An Echo middleware validates access tokens against the WorkOS JWKS for protected `/api/*` routes. The client-side WorkOS sign-in UI is deferred to the API phase — the current Astro build ships without it.

**Database.** SQLite via `modernc.org/sqlite` (pure Go, no CGO) locally; Turso (`libsql://`) in production via the pure-Go `tursodatabase/libsql-client-go` driver — `internal/db.Open` selects the driver by URL scheme, so the distroless image stays static. The Go side owns the schema: goose migrations live in `apps/api/internal/db/migrations/`, are embedded into the binary via `//go:embed`, and apply automatically on server startup.

**Email → deals → Slack.** The `apps/email_receiver` Cloudflare Worker receives forwarded newsletters via Cloudflare Email Routing and POSTs the raw RFC822 to the Go server's `POST /api/ingest` (bearer `INGEST_TOKEN`). The server parses the email, stores it in Turso, extracts deals with the Gemini API, and posts a Block Kit digest to a Slack channel (at most once per `DIGEST_INTERVAL`); mail from a sender listed in `APPROVED_FORWARDING_EMAILS` posts immediately instead of waiting, and `POST /api/admin/digest` forces one by hand. The pipeline is opt-in — it runs only when `DATABASE_URL` and `INGEST_TOKEN` are set, so the server otherwise stays a pure static host.

**Build.** `bun run build` prerenders all routes to `apps/web/dist/`; `just build` mirrors that into `apps/api/static/` so the Go binary picks it up. The Docker build runs locally and ships the prebuilt artifact in the build context — Fly's remote builder doesn't re-run `bun install`.

## Quick start

Requires [Bun](https://bun.com/), [Go](https://go.dev/) 1.26+, and [just](https://just.systems/).

```sh
# clone, then:
bun install
cp .env.local.example .env.local   # fill in WorkOS + DATABASE_URL
just dev                          # web :7001 (Astro), api :8080 (Go)
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
