import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";

type CapturedProps = {
  clientId?: string;
  apiHostname?: string;
  children?: ReactNode;
};

// Capture the props AuthProvider hands to AuthKitProvider without initializing a
// real WorkOS browser session (AuthKit would otherwise reach the network on
// mount). vi.hoisted lets the mock factory share this holder.
const captured = vi.hoisted(() => ({ props: {} as CapturedProps }));

vi.mock("@workos-inc/authkit-react", () => ({
  AuthKitProvider: (props: CapturedProps) => {
    captured.props = props;
    return props.children;
  },
}));

import AuthProvider, { safeReturnTo } from "@/integrations/workos/AuthProvider";

describe("safeReturnTo", () => {
  const origin = "https://codeselfstudy.com";

  test("keeps a same-origin path, preserving query and hash", () => {
    expect(safeReturnTo("/events/", origin)).toBe("/events/");
    expect(safeReturnTo("/events/?tab=past#top", origin)).toBe(
      "/events/?tab=past#top"
    );
  });

  test("rejects external and protocol-relative targets (open-redirect guard)", () => {
    expect(safeReturnTo("https://evil.example/phish", origin)).toBeNull();
    expect(safeReturnTo("//evil.example", origin)).toBeNull();
  });

  test("rejects empty or missing values", () => {
    expect(safeReturnTo(undefined, origin)).toBeNull();
    expect(safeReturnTo("", origin)).toBeNull();
  });
});

describe("AuthProvider", () => {
  beforeEach(() => {
    captured.props = {};
    vi.stubEnv("VITE_WORKOS_CLIENT_ID", "client_test");
    vi.stubEnv("VITE_WORKOS_API_HOSTNAME", "auth.example.com");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  test("renders children and wires WorkOS config from the environment", () => {
    render(
      <AuthProvider>
        <span>protected child</span>
      </AuthProvider>
    );

    expect(screen.getByText("protected child")).toBeInTheDocument();
    expect(captured.props.clientId).toBe("client_test");
    expect(captured.props.apiHostname).toBe("auth.example.com");
  });
});
