import { defineConfig } from "vite";
import { devtools } from "@tanstack/devtools-vite";
import { tanstackStart } from "@tanstack/react-start/plugin/vite";
import viteReact from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { nitro } from "nitro/vite";

const config = defineConfig({
  // .env.local lives at the repo root and is forwarded into the Bun process
  // via `--env-file=.env.local` in the root package.json scripts and
  // justfile. envDir would only inject VITE_* into the client build; it
  // wouldn't make non-VITE vars (DATABASE_URL, WORKOS_API_KEY) visible to
  // server-side code in env.ts. So we rely on --env-file instead.
  // Proxy API + WebSocket traffic to the Go server in dev so local routing
  // matches production (Go serves both /api and the prerendered HTML there).
  server: {
    proxy: {
      "/api": "http://localhost:8080",
      "/ws": {
        target: "ws://localhost:8080",
        ws: true,
      },
    },
  },
  resolve: {
    tsconfigPaths: true,
  },
  plugins: [
    devtools(),
    nitro({
      preset: "node",
    }),
    tailwindcss(),
    tanstackStart({
      prerender: {
        enabled: true,
        autoStaticPathsDiscovery: true,
        crawlLinks: true,
        failOnError: true,
      },
    }),
    viteReact({
      babel: {
        plugins: ["babel-plugin-react-compiler"],
      },
    }),
  ],
});

export default config;
