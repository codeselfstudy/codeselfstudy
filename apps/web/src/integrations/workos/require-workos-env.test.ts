import { afterEach, describe, expect, test, vi } from "vitest";

import { requireWorkosEnv } from "@/integrations/workos/require-workos-env";

// Invoke the integration's astro:config:setup hook the way Astro would, with a
// controllable `command`. The guard only reads `command`, so the rest of Astro's
// options object is irrelevant here.
function runSetup(command: "dev" | "build" | "preview") {
  const setup = requireWorkosEnv().hooks["astro:config:setup"];
  if (!setup) throw new Error("expected an astro:config:setup hook");
  return (setup as (opts: { command: typeof command }) => void)({ command });
}

describe("requireWorkosEnv", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  describe("on build", () => {
    test("passes when both WorkOS vars are set", () => {
      vi.stubEnv("VITE_WORKOS_CLIENT_ID", "client_live");
      vi.stubEnv("VITE_WORKOS_API_HOSTNAME", "auth.example.com");
      expect(() => runSetup("build")).not.toThrow();
    });

    test("throws when the client id is empty", () => {
      vi.stubEnv("VITE_WORKOS_CLIENT_ID", "");
      vi.stubEnv("VITE_WORKOS_API_HOSTNAME", "auth.example.com");
      expect(() => runSetup("build")).toThrow(/VITE_WORKOS_CLIENT_ID/);
    });

    test("treats a whitespace-only value as empty", () => {
      vi.stubEnv("VITE_WORKOS_CLIENT_ID", "   ");
      vi.stubEnv("VITE_WORKOS_API_HOSTNAME", "auth.example.com");
      expect(() => runSetup("build")).toThrow(/VITE_WORKOS_CLIENT_ID/);
    });

    test("throws when the api hostname is empty", () => {
      vi.stubEnv("VITE_WORKOS_CLIENT_ID", "client_live");
      vi.stubEnv("VITE_WORKOS_API_HOSTNAME", "");
      expect(() => runSetup("build")).toThrow(/VITE_WORKOS_API_HOSTNAME/);
    });

    test("names every missing var in one message", () => {
      vi.stubEnv("VITE_WORKOS_CLIENT_ID", "");
      vi.stubEnv("VITE_WORKOS_API_HOSTNAME", "");
      expect(() => runSetup("build")).toThrow(
        /VITE_WORKOS_CLIENT_ID, VITE_WORKOS_API_HOSTNAME/
      );
    });
  });

  test("leaves the dev server alone even when vars are missing", () => {
    vi.stubEnv("VITE_WORKOS_CLIENT_ID", "");
    vi.stubEnv("VITE_WORKOS_API_HOSTNAME", "");
    expect(() => runSetup("dev")).not.toThrow();
  });
});
