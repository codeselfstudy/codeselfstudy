// @ts-check
import { defineConfig } from "astro/config";

import react from "@astrojs/react";

import tailwindcss from "@tailwindcss/vite";

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

  integrations: [react()],

  vite: {
    plugins: [tailwindcss()],
  },
});
