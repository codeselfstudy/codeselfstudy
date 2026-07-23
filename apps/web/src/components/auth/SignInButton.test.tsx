import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import SignInButton from "@/components/auth/SignInButton";

// The cookie-gated /api/me is the single source of auth truth: 200 -> signed in,
// 401 -> signed out. Drive both states through a mocked fetch, and capture
// window.location.assign to assert the /auth navigations.

function meOk(data: { email?: string; name?: string; avatar?: string }) {
  return { ok: true, status: 200, json: async () => data };
}
const meUnauthorized = { ok: false, status: 401, json: async () => null };

const realLocation = window.location;
let assignMock: ReturnType<typeof vi.fn>;

beforeEach(() => {
  assignMock = vi.fn();
  Object.defineProperty(window, "location", {
    configurable: true,
    writable: true,
    value: { pathname: "/events/", search: "", hash: "", assign: assignMock },
  });
});

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
  Object.defineProperty(window, "location", {
    configurable: true,
    writable: true,
    value: realLocation,
  });
});

describe("SignInButton", () => {
  test("signed out: shows an enabled Sign In once /api/me returns 401", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(meUnauthorized));
    render(<SignInButton />);

    expect(
      await screen.findByRole("button", { name: "Sign In" })
    ).toBeEnabled();
    expect(
      screen.queryByRole("button", { name: "Sign Out" })
    ).not.toBeInTheDocument();
  });

  test("signed in: shows the email from /api/me and a Sign Out button", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(
          meOk({ email: "ada@example.com", name: "Ada Lovelace" })
        )
    );
    render(<SignInButton />);

    expect(await screen.findByText("ada@example.com")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Sign Out" })
    ).toBeInTheDocument();
  });

  test("requests /api/me with same-origin credentials so the cookie is sent", async () => {
    const fetchMock = vi.fn().mockResolvedValue(meUnauthorized);
    vi.stubGlobal("fetch", fetchMock);
    render(<SignInButton />);

    await screen.findByRole("button", { name: "Sign In" });
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/me",
      expect.objectContaining({ credentials: "same-origin" })
    );
  });

  test("clicking Sign In navigates to /auth/login with the current path as returnTo", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(meUnauthorized));
    const user = userEvent.setup();
    render(<SignInButton />);

    await user.click(await screen.findByRole("button", { name: "Sign In" }));

    expect(assignMock).toHaveBeenCalledWith(
      "/auth/login?returnTo=%2Fevents%2F"
    );
  });

  test("returnTo preserves the query string and hash, not just the path", async () => {
    Object.defineProperty(window, "location", {
      configurable: true,
      writable: true,
      value: {
        pathname: "/events/",
        search: "?tab=upcoming",
        hash: "#speakers",
        assign: assignMock,
      },
    });
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(meUnauthorized));
    const user = userEvent.setup();
    render(<SignInButton />);

    await user.click(await screen.findByRole("button", { name: "Sign In" }));

    expect(assignMock).toHaveBeenCalledWith(
      `/auth/login?returnTo=${encodeURIComponent("/events/?tab=upcoming#speakers")}`
    );
  });

  test("clicking Sign Out navigates to /auth/logout", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(meOk({ email: "ada@example.com" }))
    );
    const user = userEvent.setup();
    render(<SignInButton />);

    await user.click(await screen.findByRole("button", { name: "Sign Out" }));

    expect(assignMock).toHaveBeenCalledWith(
      "/auth/logout?returnTo=%2Fevents%2F"
    );
  });
});
