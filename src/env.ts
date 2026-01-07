import { createEnv } from "@t3-oss/env-core";
import { z } from "zod";

export const env = createEnv({
  server: {
    SERVER_URL: z.url().optional(), // Make optional or ensure it's set in Prod
    WORKOS_API_KEY: z.string(),
    DATABASE_URL: z.string(),
    TURSO_AUTH_TOKEN: z.string().optional(),
  },

  /**
   * The prefix that client-side variables must have. This is enforced both at
   * a type-level and at runtime.
   */
  clientPrefix: "VITE_",

  client: {
    VITE_WORKOS_CLIENT_ID: z.string().min(1),
    VITE_WORKOS_API_HOSTNAME: z.string().min(1),
  },

  /**
   * What object holds the environment variables at runtime. This is usually
   * `process.env` or `import.meta.env`.
   */
  runtimeEnv:
    typeof window === "undefined"
      ? process.env
      : { ...process.env, ...(import.meta as any).env },

  /**
   * By default, this library will feed the environment variables directly to
   * the Zod validator.
   *
   * This means that if you have an empty string for a value that is supposed
   * to be a number (e.g. `PORT=` in a ".env" file), Zod will incorrectly flag
   * it as a type mismatch violation. Additionally, if you have an empty string
   * for a value that is supposed to be a string with a default value (e.g.
   * `DOMAIN=` in an ".env" file), the default value will never be applied.
   *
   * In order to solve these issues, we recommend that all new projects
   * explicitly specify this option as true.
   */
  emptyStringAsUndefined: true,
});
