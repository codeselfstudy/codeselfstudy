import type { AstroIntegration } from "astro";

// Client-exposed WorkOS config that the browser bundle cannot work without.
// These mirror the two vars declared in env.d.ts and consumed by AuthProvider.tsx.
const REQUIRED_VARS = [
  "PUBLIC_WORKOS_CLIENT_ID",
  "PUBLIC_WORKOS_API_HOSTNAME",
] as const;

// Build-time guard for the client-exposed WorkOS config.
//
// The web app is built locally and its `dist/` ships prebuilt into the Fly image
// (see Dockerfile + the `just deploy` recipe). Vite *statically inlines*
// `import.meta.env.PUBLIC_WORKOS_CLIENT_ID` — read in AuthProvider.tsx — at build
// time, so if that var is empty when `astro build` runs, the shipped bundle bakes
// in `undefined`. At runtime AuthKit then calls `createClient(undefined)` and
// throws `NoClientIdProvidedException`, which takes down the whole navbar island
// (Navbar.tsx wraps the entire nav in <AuthProvider>). That failure is invisible
// until the broken bundle is already in production.
//
// This integration turns that silent, ship-to-prod failure into a loud, local
// build failure: on `astro build` it aborts when either required var is empty.
// It deliberately does nothing on `dev`/`preview` so pure-UI local work isn't
// blocked by missing auth config.
//
// We read `process.env` directly (rather than Vite's `loadEnv`) because that is
// exactly what the deploy path feeds Vite: `just deploy` runs
// `bun run --env-file=.env.local --filter web build`, which injects the vars
// into `process.env` for `astro build` to inline. Importing `vite` here would
// also break config loading in CI, where `vite` is only a transitive dep and
// does not resolve from a file under `src/`. (The repo keeps no `apps/web/.env*`
// files, so there is no file-only source of these vars to miss.)
export function requireWorkosEnv(): AstroIntegration {
  return {
    name: "require-workos-env",
    hooks: {
      "astro:config:setup": ({ command }) => {
        if (command !== "build") return; // leave `dev` / `preview` alone
        const missing = REQUIRED_VARS.filter(
          (key) => !process.env[key]?.trim()
        );
        if (missing.length > 0) {
          throw new Error(
            `Refusing to build: ${missing.join(", ")} ` +
              `${missing.length > 1 ? "are" : "is"} empty. Set ` +
              `${missing.length > 1 ? "them" : "it"} in .env.local ` +
              "(see .env.local.example) before `just build` / `just deploy`. Building " +
              "without it inlines `undefined`, shipping a navbar that throws " +
              "NoClientIdProvidedException at runtime."
          );
        }
      },
    },
  };
}
