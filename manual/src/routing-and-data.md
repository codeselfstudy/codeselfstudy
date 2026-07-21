# Routing & Data

> All paths in this section are relative to `apps/web/`.

## Routing (Astro)

Routes are file-based in `src/pages/`. Each `.astro` file becomes a page: `src/pages/about.astro` builds to `dist/about/index.html` and is served at `/about/`. Add a route by adding a file.

The site is statically generated (`output: "static"`) and configured with `trailingSlash: "always"` and `build.format: "directory"` in `astro.config.mjs`, so every page is a directory with an `index.html` and every canonical URL ends in a slash. Because of that, **all internal links must carry a trailing slash** (see AGENTS.md).

```astro
<a href="/about/">About</a>
```

There is no client-side router — this is a multi-page app. Each navigation is a full page load, so links are plain `<a>` tags, not a framework `Link` component.

## Layout

Every page renders inside `src/layouts/Layout.astro`, which owns the `<html>` document: the `<head>` (title, description, canonical URL, Open Graph tags via `src/lib/metadata.ts`, and production-only analytics), the navbar, and the footer. Content pages import `Layout` and pass their metadata as props:

```astro
---
import Layout from "@/layouts/Layout.astro";
import PageWrapper from "@/components/PageWrapper.astro";
---

<Layout title="About" description="About the Code Self Study website">
  <PageWrapper>
    <h1>About Us</h1>
    <p>…</p>
  </PageWrapper>
</Layout>
```

`title` and `description` are required props, so `astro check` (run as part of `bun run build`) fails any page that forgets its metadata. `PageWrapper.astro` is the shared content container; the homepage skips it to render full-bleed.

## Interactivity (islands)

Pages are static HTML by default and ship no JavaScript. When a component needs to run in the browser it becomes an **island** — a React component rendered with a `client:*` directive. The navbar is the only island today; its mobile drawer needs client-side state, so `Layout.astro` renders it with `client:load`:

```astro
---
import Navbar from "@/components/Navbar.tsx";
---

<Navbar client:load />
```

Keep islands small: everything outside them stays zero-JS.

## Data fetching

The current site fetches no data — every page is prerendered at build time. When the Go API surface is built out, calls to `/api/*` will happen from islands (or future server code), not from the static pages. There are no route loaders.

## Redirects and URL canonicalization

Because a static build can't emit real HTTP redirects, the **Go server owns the URL layer**:

- **Legacy redirects** (`apps/api/redirects.go`) map old paths to their new homes with 308s (e.g. `/book` → `/learn/`, `/blog/*` → `/learn/`).
- **Trailing-slash canonicalization** (`apps/api/main.go`) 301s a non-slashed page URL to its slashed form, so each page has a single canonical URL.

Neither lives in the frontend. See [Architecture](./architecture.md) for the full request flow.
