// @ts-check
import js from "@eslint/js";
import tseslint from "typescript-eslint";
import eslintPluginAstro from "eslint-plugin-astro";
import reactHooks from "eslint-plugin-react-hooks";

// Minimal flat config for the Astro app: JS/TS recommended rules, the Astro
// plugin for `.astro` files, and React Hooks rules for the interactive islands
// (the navbar). The old app's bespoke `enforce-route-metadata` rule is not
// ported — that discipline is enforced structurally instead: `Layout.astro`
// declares `title`/`description` as required props and `astro check` (run as
// part of the build script) fails any page that omits them.
export default [
  {
    ignores: ["dist/**", "coverage/**", ".astro/**"],
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...eslintPluginAstro.configs.recommended,
  {
    files: ["**/*.{ts,tsx}"],
    plugins: { "react-hooks": reactHooks },
    rules: {
      "react-hooks/rules-of-hooks": "error",
      "react-hooks/exhaustive-deps": "warn",
    },
  },
];
