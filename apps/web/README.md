# Code Self Study — web

The Astro frontend of the Code Self Study website: statically generated pages with React islands, prerendered to `dist/` and served by the Go backend (`apps/api/`) in production.

Work from the repo root:

```sh
bun install
just dev       # Astro dev server on http://localhost:7001 (+ Go API on :8080)
just build     # prerender to dist/ and mirror into apps/api/static/
just test      # full test suite (Go + web + Worker)
```

See the [root README](../../README.md) and the [manual](../../manual/src/SUMMARY.md) for architecture, conventions, and deployment.
