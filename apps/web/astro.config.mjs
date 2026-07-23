// @ts-check
import { defineConfig } from "astro/config";

import react from "@astrojs/react";
import sitemap from "@astrojs/sitemap";

import tailwindcss from "@tailwindcss/vite";

import { requireWorkosEnv } from "./src/integrations/workos/require-workos-env";

// https://astro.build/config
export default defineConfig({
  site: "https://codeselfstudy.com",

  // Statically generated site served by the Go/Echo server, whose static
  // handler resolves `<path>/index.html`. Trailing slashes are mandatory on
  // internal links (see AGENTS.md); pin both settings explicitly so the URL
  // contract with the Go server never drifts on an Astro default change.
  trailingSlash: "always",
  build: {
    format: "directory",
  },

  integrations: [
    react(),
    // Fail `astro build` loudly when the client-exposed WorkOS vars are empty,
    // instead of inlining `undefined` and shipping a navbar that throws
    // NoClientIdProvidedException at runtime. No-ops on `dev`/`preview`.
    requireWorkosEnv(),
    // Exclude /s/ (the Slack invite page, which is noindex) from the sitemap.
    sitemap({
      filter: (page) => page !== "https://codeselfstudy.com/s/",
    }),
  ],

  vite: {
    // Astro defaults Vite's envPrefix to `PUBLIC_`, so only PUBLIC_-prefixed
    // vars are inlined into client bundles. Our WorkOS client config is exposed
    // under the `VITE_` prefix (ported from the old Vite/TanStack app, and
    // shared with the Go API's VITE_WORKOS_* fallback in apps/api/main.go), so
    // `VITE_` must be added here — otherwise import.meta.env.VITE_WORKOS_CLIENT_ID
    // is `undefined` in the browser and AuthKit throws NoClientIdProvidedException.
    envPrefix: ["PUBLIC_", "VITE_"],
    plugins: [tailwindcss()],
  },
});
