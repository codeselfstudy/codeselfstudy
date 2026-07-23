/// <reference types="vitest/config" />
import { getViteConfig } from "astro/config";

// Use Astro's Vite config for tests so path aliases (`@/*`) and integration
// plugins (React) resolve identically to the app build. Coverage discipline:
// 100% on `src/lib/**` (pure, testable logic), and a 90% floor everywhere else
// so a new untested component can't quietly drag the suite back down. CI runs
// `bun run --filter web test:coverage`, so these thresholds are the gate.
// Note: files matched by the `src/lib/**` glob are checked against that entry
// and excluded from the global numbers. Only `.ts`/`.tsx` is measured — `.astro`
// pages, components and layouts are outside `include`.
export default getViteConfig({
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./test/setup.ts"],
    coverage: {
      provider: "v8",
      reporter: ["text", "html", "lcov", "json-summary"],
      reportsDirectory: "./coverage",
      include: ["src/**/*.{ts,tsx}"],
      exclude: ["src/**/*.test.{ts,tsx}", "src/**/*.d.ts"],
      thresholds: {
        lines: 90,
        statements: 90,
        functions: 90,
        branches: 90,
        "src/lib/**": {
          lines: 100,
          branches: 100,
          functions: 100,
          statements: 100,
        },
      },
    },
  },
});
