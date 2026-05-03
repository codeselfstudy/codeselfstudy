import { defineConfig } from "drizzle-kit";
import { env } from "@/env";

const { DATABASE_URL, TURSO_AUTH_TOKEN } = env;

export default defineConfig({
  out: "./drizzle",
  schema: "./src/db/schema.ts",
  dialect: "sqlite",
  dbCredentials: {
    url: DATABASE_URL,
    token: TURSO_AUTH_TOKEN,
  },
});
