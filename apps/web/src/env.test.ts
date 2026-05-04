import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";

const REQUIRED = {
  WORKOS_API_KEY: "test_workos_api_key",
  VITE_WORKOS_CLIENT_ID: "client_test",
  VITE_WORKOS_API_HOSTNAME: "auth.example.com",
} as const;

function stubAllRequired() {
  for (const [key, value] of Object.entries(REQUIRED)) {
    vi.stubEnv(key, value);
  }
}

beforeEach(() => {
  vi.resetModules();
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
  stubAllRequired();
  vi.stubEnv("SERVER_URL", "");
});

afterEach(() => {
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
});

describe("env (client/browser branch — window defined)", () => {
  test("parses required server vars exposed to the module", async () => {
    expect(typeof window).toBe("object");
    const { env } = await import("./env");
    expect(env.VITE_WORKOS_CLIENT_ID).toBe(REQUIRED.VITE_WORKOS_CLIENT_ID);
    expect(env.VITE_WORKOS_API_HOSTNAME).toBe(
      REQUIRED.VITE_WORKOS_API_HOSTNAME
    );
  });

  test("blocks server-only var access at runtime in the browser", async () => {
    const { env } = await import("./env");
    expect(() => (env as { WORKOS_API_KEY: string }).WORKOS_API_KEY).toThrow(
      /server-side environment variable/
    );
  });
});

describe("env (server branch — window undefined)", () => {
  beforeEach(() => {
    vi.stubGlobal("window", undefined);
  });

  test("reads server vars directly from process.env", async () => {
    expect(typeof window).toBe("undefined");
    const { env } = await import("./env");
    expect(env.WORKOS_API_KEY).toBe(REQUIRED.WORKOS_API_KEY);
  });

  test("treats empty SERVER_URL as undefined", async () => {
    const { env } = await import("./env");
    expect(env.SERVER_URL).toBeUndefined();
  });

  test("accepts a valid SERVER_URL", async () => {
    vi.stubEnv("SERVER_URL", "https://example.com");
    const { env } = await import("./env");
    expect(env.SERVER_URL).toBe("https://example.com");
  });

  test("throws when WORKOS_API_KEY is missing", async () => {
    vi.stubEnv("WORKOS_API_KEY", "");
    await expect(import("./env")).rejects.toThrow();
  });

  test("throws when VITE_WORKOS_CLIENT_ID is missing", async () => {
    vi.stubEnv("VITE_WORKOS_CLIENT_ID", "");
    await expect(import("./env")).rejects.toThrow();
  });

  test("throws when VITE_WORKOS_API_HOSTNAME is missing", async () => {
    vi.stubEnv("VITE_WORKOS_API_HOSTNAME", "");
    await expect(import("./env")).rejects.toThrow();
  });
});
