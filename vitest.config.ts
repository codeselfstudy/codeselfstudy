import { mergeConfig } from "vitest/config";
import viteReact from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

const viteConfig = {
  resolve: {
    tsconfigPaths: true,
  },
  plugins: [
    tailwindcss(),
    viteReact({
      babel: {
        plugins: ["babel-plugin-react-compiler"],
      },
    }),
  ],
};

export default mergeConfig(viteConfig, {
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./test/setup.ts"],
    css: true,
    coverage: {
      provider: "v8",
      reporter: ["text", "html", "lcov", "json-summary"],
      reportsDirectory: "./coverage",
      include: ["src/**/*.{ts,tsx}"],
      exclude: [
        "src/**/*.test.{ts,tsx}",
        "src/**/*.d.ts",
        "src/routeTree.gen.ts",
        "src/router.tsx",
      ],
      thresholds: {
        lines: 0,
        branches: 0,
        functions: 0,
        statements: 0,
        "src/lib/**": {
          lines: 100,
          branches: 100,
          functions: 100,
          statements: 100,
        },
        "src/data/**": {
          lines: 100,
          branches: 100,
          functions: 100,
          statements: 100,
        },
        "src/hooks/**": {
          lines: 100,
          branches: 100,
          functions: 100,
          statements: 100,
        },
        "src/env.ts": {
          lines: 100,
          branches: 66,
          functions: 100,
          statements: 100,
        },
      },
    },
  },
});
