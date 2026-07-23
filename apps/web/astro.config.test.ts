import { describe, expect, it } from "vitest";

import astroConfig from "./astro.config.mjs";

// Regression guard for the WorkOS "Missing Client ID" bug.
//
// Astro inlines into client bundles only the env vars matching Vite's
// `envPrefix`, and it defaults that prefix to `PUBLIC_`. Our WorkOS client
// config is exposed under the `VITE_` prefix (see AuthProvider.tsx and the Go
// API's VITE_WORKOS_* fallback in apps/api/main.go), so the config widens
// `envPrefix` to include `VITE_`. If a future edit drops it,
// `import.meta.env.VITE_WORKOS_CLIENT_ID` silently becomes `undefined` in the
// browser and AuthKit throws NoClientIdProvidedException in production — a
// failure the build-time env guard cannot catch, because it reads `process.env`
// (which is populated) rather than the inlined client bundle (which is not).
describe("astro.config vite.envPrefix", () => {
  it("inlines both PUBLIC_ and VITE_ prefixed vars into client bundles", () => {
    const envPrefix = astroConfig.vite?.envPrefix;
    expect(envPrefix).toContain("PUBLIC_");
    expect(envPrefix).toContain("VITE_");
  });
});
