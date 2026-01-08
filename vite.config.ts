import { defineConfig } from "vite";
import { devtools } from "@tanstack/devtools-vite";
import { tanstackStart } from "@tanstack/react-start/plugin/vite";
import viteReact from "@vitejs/plugin-react";
import viteTsConfigPaths from "vite-tsconfig-paths";
import tailwindcss from "@tailwindcss/vite";
import { nitro } from "nitro/vite";

const config = defineConfig({
  server: {
    proxy: {
      "/ws": {
        target: "ws://localhost:8081",
        ws: true,
        changeOrigin: true,
      },
    },
  },

  plugins: [
    devtools(),
    nitro({
      preset: "bun",
      // TODO: is this right for TanStack Start? It already has a server.
      serverDir: "src/server",
    }),
    // this is the plugin that enables path aliases
    viteTsConfigPaths({
      projects: ["./tsconfig.json"],
    }),
    tailwindcss(),
    tanstackStart(),
    viteReact({
      babel: {
        plugins: ["babel-plugin-react-compiler"],
      },
    }),
  ],
});

export default config;
