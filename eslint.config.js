//  @ts-check

import { tanstackConfig } from "@tanstack/eslint-config";

export default [
  ...tanstackConfig,
  {
    ignores: [
      "TANK/**",
      "**/*.bak/**",
      "**/*.min.js",
      "manual/mermaid-init.js",
      "manual/book/**",
    ],
  },
];
