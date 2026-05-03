import { createClient } from "@libsql/client";
import { drizzle } from "drizzle-orm/libsql";
import * as schema from "./schema";
import { env } from "@/env";

const client = (() => {
  try {
    return createClient({
      url: env.DATABASE_URL,
      authToken: env.TURSO_AUTH_TOKEN,
    });
  } catch (error) {
    console.error(
      "Failed to initialize libsql client. Please verify DATABASE_URL and TURSO_AUTH_TOKEN.",
      error
    );
    throw new Error("libsql client initialization failed.", {
      cause: error as Error,
    });
  }
})();

export const db = drizzle(client, { schema });
