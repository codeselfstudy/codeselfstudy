import type { AstroIntegration } from "astro";
import { loadEnv } from "vite";

// Client-exposed WorkOS config that the browser bundle cannot work without.
// These mirror the two vars declared in env.d.ts and consumed by AuthProvider.tsx.
const REQUIRED_VARS = [
  "VITE_WORKOS_CLIENT_ID",
  "VITE_WORKOS_API_HOSTNAME",
] as const;

// Build-time guard for the client-exposed WorkOS config.
//
// The web app is built locally and its `dist/` ships prebuilt into the Fly image
// (see Dockerfile + the `just deploy` recipe). Vite *statically inlines*
// `import.meta.env.VITE_WORKOS_CLIENT_ID` — read in AuthProvider.tsx — at build
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
// We read through Vite's `loadEnv` rather than `process.env` directly so the
// check sees exactly what Vite will inline: the `VITE_`-prefixed vars that
// `bun run --env-file=.env.local` injects into `process.env`, plus any
// `apps/web/.env*` files.
export function requireWorkosEnv(): AstroIntegration {
  return {
    name: "require-workos-env",
    hooks: {
      "astro:config:setup": ({ command }) => {
        if (command !== "build") return; // leave `dev` / `preview` alone
        const env = loadEnv("production", process.cwd(), "VITE_");
        const missing = REQUIRED_VARS.filter((key) => !env[key]?.trim());
        if (missing.length > 0) {
          throw new Error(
            `Refusing to build: ${missing.join(", ")} ` +
              `${missing.length > 1 ? "are" : "is"} empty. Set ` +
              `${missing.length > 1 ? "them" : "it"} in .env.local ` +
              "(see .env-example) before `just build` / `just deploy`. Building " +
              "without it inlines `undefined`, shipping a navbar that throws " +
              "NoClientIdProvidedException at runtime."
          );
        }
      },
    },
  };
}
