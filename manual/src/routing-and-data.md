# Routing & Data

> All paths in this section are relative to `apps/web/`.

## Routing (Astro)

Routes are file-based in `src/pages/`. Each `.astro` file becomes a page: `src/pages/about.astro` builds to `dist/about/index.html` and is served at `/about/`. Add a route by adding a file.

The site is statically generated (static output is Astro's default) and configured with `trailingSlash: "always"` and `build.format: "directory"` in `astro.config.mjs`, so every page is a directory with an `index.html` and every canonical URL ends in a slash. Because of that, **all internal links must carry a trailing slash** (see AGENTS.md).

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

Pages are static HTML by default and ship no JavaScript. When a component needs to run in the browser it becomes an **island** — a React component rendered with a `client:*` directive. The main island is the navbar; it owns client-side state (the mobile drawer) and the auth controls (`SignInButton` / `UserMenu`), which read the `css_auth` hint cookie and a cached profile so the right control paints on the first frame. `Layout.astro` renders it browser-only with `client:only="react"`:

```astro
---
import Navbar from "@/components/Navbar.tsx";
---

<Navbar client:only="react" />
```

Keep islands small: everything outside them stays zero-JS.

## Data fetching

Every page is prerendered at build time. The one exception is the navbar island: it reads `/api/me` as a plain same-origin fetch — the browser sends the first-party session cookie, so no token handling happens in JavaScript. Signing in is a full-page navigation to the Go server's `/auth/login`. Static pages fetch nothing, and there are no route loaders.

## Redirects and URL canonicalization

Because a static build can't emit real HTTP redirects, the **Go server owns the URL layer**:

- **Legacy redirects** (`apps/api/redirects.go`) map old paths to their new homes with 308s (e.g. `/book` → `/learn/`, `/blog/*` → `/learn/`).
- **Trailing-slash canonicalization** (`apps/api/main.go`) 301s a non-slashed page URL to its slashed form, so each page has a single canonical URL.

Neither lives in the frontend. See [Architecture](./architecture.md) for the full request flow.
