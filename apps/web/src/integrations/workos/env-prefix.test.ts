import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

// Regression guard for the WorkOS "Missing Client ID" saga. The client-exposed
// WorkOS vars MUST use Astro's `PUBLIC_` prefix — the only prefix Astro inlines
// into the browser bundle. If they regress to `VITE_`, `import.meta.env.*` is
// `undefined` at runtime and AuthKit throws NoClientIdProvidedException. The
// build-time guard (require-workos-env) can't catch that: it reads `process.env`
// (populated by --env-file), not the inlined client bundle. So assert the source
// of the two files that consume/validate the vars.
const read = (relative: string) =>
  readFileSync(new URL(relative, import.meta.url), "utf8");

describe("WorkOS client env vars use Astro's PUBLIC_ prefix", () => {
  for (const file of [
    "./AuthProvider.tsx",
    "./require-workos-env.ts",
  ] as const) {
    it(`${file} references PUBLIC_WORKOS_* and no VITE_WORKOS_*`, () => {
      const source = read(file);
      expect(source).toMatch(/PUBLIC_WORKOS_CLIENT_ID/);
      expect(source).toMatch(/PUBLIC_WORKOS_API_HOSTNAME/);
      expect(source).not.toMatch(/VITE_WORKOS_/);
    });
  }
});
