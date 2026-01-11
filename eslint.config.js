//  @ts-check

import { tanstackConfig } from "@tanstack/eslint-config";

/**
 * Custom ESLint plugin to enforce page metadata.
 */
const routeMetadataPlugin = {
  rules: {
    "enforce-route-metadata": {
      meta: {
        type: "problem",
        docs: {
          description:
            "Enforce 'head' property in createFileRoute configuration",
        },
        schema: [],
      },
      /**
       * @param {import('eslint').Rule.RuleContext} context
       * @returns {import('eslint').Rule.RuleListener}
       */
      create(context) {
        return {
          /** @param {any} node */
          CallExpression(node) {
            // Check for createFileRoute()({...}) pattern
            if (
              node.callee.type === "CallExpression" &&
              node.callee.callee.type === "Identifier" &&
              node.callee.callee.name === "createFileRoute"
            ) {
              const configArg = node.arguments[0];
              if (configArg && configArg.type === "ObjectExpression") {
                const hasHead = configArg.properties.some(
                  /** @param {any} prop */
                  (prop) =>
                    prop.type === "Property" &&
                    prop.key.type === "Identifier" &&
                    prop.key.name === "head"
                );

                if (!hasHead) {
                  context.report({
                    node,
                    message:
                      "Routes must define a 'head' property for metadata using createMetadata.",
                  });
                }
              }
            }
          },
        };
      },
    },
  },
};

export default [
  ...tanstackConfig,
  {
    plugins: {
      "local-rules": routeMetadataPlugin,
    },
    rules: {
      "local-rules/enforce-route-metadata": "error",
    },
  },
  {
    ignores: [
      ".output/**",
      "dist/**",
      "TANK/**",
      "**/*.bak/**",
      "**/*.min.js",
      "manual/mermaid-init.js",
      "manual/book/**",
    ],
  },
];
