import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import SignInButton from "@/components/auth/SignInButton";

// The cookie-gated /api/me is the single source of auth truth: 200 -> signed in,
// 401 -> signed out. Drive both states through a mocked fetch, and capture
// window.location.assign to assert the /auth navigations. When signed in the
// control renders UserMenu, so the username shows in the trigger and Sign Out
// lives inside the opened menu (never the email — see #351).

function meOk(data: { email?: string; username?: string; avatar?: string }) {
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
      screen.queryByRole("button", { name: /account menu/i })
    ).not.toBeInTheDocument();
  });

  test("signed in: shows the username, not the email, with Sign Out in the menu", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(
          meOk({ email: "ada@example.com", username: "adalovelace" })
        )
    );
    const user = userEvent.setup();
    render(<SignInButton />);

    expect(await screen.findByText("adalovelace")).toBeInTheDocument();
    expect(screen.queryByText("ada@example.com")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /account menu/i }));
    expect(
      await screen.findByRole("menuitem", { name: "Sign Out" })
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

  test("Sign Out in the menu navigates to /auth/logout with returnTo", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(meOk({ username: "adalovelace" }))
    );
    const user = userEvent.setup();
    render(<SignInButton />);

    await user.click(
      await screen.findByRole("button", { name: /account menu/i })
    );
    await user.click(await screen.findByRole("menuitem", { name: "Sign Out" }));

    expect(assignMock).toHaveBeenCalledWith(
      "/auth/logout?returnTo=%2Fevents%2F"
    );
  });

  test("a failed /api/me request falls back to signed out", async () => {
    // The API being down must not take the whole navbar with it.
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));
    render(<SignInButton />);

    expect(
      await screen.findByRole("button", { name: "Sign In" })
    ).toBeEnabled();
  });

  test("renders the avatar in the trigger when /api/me returns one", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        meOk({
          username: "adalovelace",
          avatar: "https://example.com/ada.png",
        })
      )
    );
    const { container } = render(<SignInButton />);

    await screen.findByRole("button", { name: /account menu/i });
    // The avatar is decorative (alt=""), so it is not in the a11y tree.
    expect(container.querySelector("img")).toHaveAttribute(
      "src",
      "https://example.com/ada.png"
    );
  });

  test("falls back to the email label when /api/me returns no username", async () => {
    // Degrades sensibly when the server runs without a database.
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(meOk({ email: "ada@example.com" }))
    );
    render(<SignInButton />);

    expect(
      await screen.findByRole("button", {
        name: "Account menu for ada@example.com",
      })
    ).toBeInTheDocument();
  });

  test("signed in with an empty profile still renders the account menu", async () => {
    // A 200 with no username, email or avatar is still a valid session; render
    // the menu trigger (named for assistive tech) without an empty label or a
    // broken image.
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(meOk({})));
    const user = userEvent.setup();
    const { container } = render(<SignInButton />);

    const trigger = await screen.findByRole("button", { name: "Account menu" });
    expect(trigger).toBeInTheDocument();
    expect(container.querySelector("img")).toBeNull();

    await user.click(trigger);
    expect(
      await screen.findByRole("menuitem", { name: "Sign Out" })
    ).toBeInTheDocument();
  });

  test("ignores a /api/me response that arrives after unmount", async () => {
    let settle: (value: unknown) => void = () => {};
    const pending = new Promise((resolve) => {
      settle = resolve;
    });
    vi.stubGlobal("fetch", vi.fn().mockReturnValue(pending));
    const consoleError = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});

    const { unmount } = render(<SignInButton />);
    unmount();

    // The `cancelled` flag set by the effect cleanup is what keeps this late
    // resolution from touching state on a component that is already gone.
    settle(meOk({ username: "adalovelace" }));
    await pending;
    await Promise.resolve();
    await Promise.resolve();

    expect(consoleError).not.toHaveBeenCalled();
    expect(
      screen.queryByRole("button", { name: /account menu/i })
    ).not.toBeInTheDocument();
  });

  test("ignores a /api/me failure that arrives after unmount", async () => {
    let fail: (reason: unknown) => void = () => {};
    const pending = new Promise((_resolve, reject) => {
      fail = reject;
    });
    vi.stubGlobal("fetch", vi.fn().mockReturnValue(pending));
    const consoleError = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});

    const { unmount } = render(<SignInButton />);
    unmount();

    // Same guard as above, on the error path: the component's own .catch keeps
    // the rejection from surfacing, and `cancelled` keeps it from setting state.
    fail(new Error("offline"));
    await pending.catch(() => {});
    await Promise.resolve();
    await Promise.resolve();

    expect(consoleError).not.toHaveBeenCalled();
    expect(
      screen.queryByRole("button", { name: "Sign In" })
    ).not.toBeInTheDocument();
  });
});
