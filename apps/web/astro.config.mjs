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
    plugins: [tailwindcss()],
  },
});
