import { createClient } from "@libsql/client";
import { drizzle } from "drizzle-orm/libsql";
import * as schema from "./schema";
import { env } from "@/env";

// Resolve absolute path for local SQLite files to avoid CWD issues in different environments
const getDbUrl = (url: string) => {
  if (url.includes("://") || url.startsWith("file:")) return url;
  return `file:${process.cwd()}/${url}`;
};

const client = createClient({
  url: getDbUrl(env.DATABASE_URL),
  authToken: env.TURSO_AUTH_TOKEN,
});

export const db = drizzle(client, { schema });
