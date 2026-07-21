// @ts-check
import js from "@eslint/js";
import tseslint from "typescript-eslint";

// Minimal flat config for the Cloudflare Worker (TypeScript, no DOM/React): the
// JS + TS recommended rule sets. Shipped per-app on purpose: the repo has no
// root eslint config, and the pre-push hook runs `eslint` on pushed files by
// walking up from each file's directory — without a config here, pushing the
// worker's .ts would fail to resolve one.
export default [
  {
    ignores: ["dist/**"],
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ["**/*.ts"],
    rules: {
      // Allow intentionally-unused args/vars prefixed with _ (e.g. the unused
      // ExecutionContext in the email() handler).
      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
      ],
    },
  },
  {
    // Pin this config's TSConfig root so typescript-eslint doesn't error on
    // "multiple candidate TSConfigRootDirs" now that the repo has more than one
    // app tsconfig (apps/web, apps/email_receiver).
    languageOptions: {
      parserOptions: {
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
];
