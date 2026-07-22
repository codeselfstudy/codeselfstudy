/// <reference types="vitest/config" />
import { getViteConfig } from "astro/config";

// Use Astro's Vite config for tests so path aliases (`@/*`) and integration
// plugins (React) resolve identically to the app build. Coverage discipline is
// ported from the old app: 100% on `src/lib/**` (pure, testable logic), and no
// global floor so component/page files don't force meaningless test coverage.
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
