import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import Navbar from "@/components/Navbar";
import { AUTH_FLAG_COOKIE, writeAccountHint } from "@/lib/authHint";

// Navbar renders SignInButton, which fetches the cookie-gated /api/me on mount.
// Stub fetch as signed-out (401) so the nav tests never hit the network (jsdom
// ships no fetch of its own); a single SignInButton renders in the top bar.
beforeEach(() => {
  vi.stubGlobal(
    "fetch",
    vi
      .fn()
      .mockResolvedValue({ ok: false, status: 401, json: async () => null })
  );
});

afterEach(() => {
  vi.unstubAllGlobals();
  localStorage.clear();
  document.cookie = `${AUTH_FLAG_COOKIE}=; max-age=0`;
});

describe("Navbar", () => {
  test("renders the logo and desktop links with trailing-slash hrefs", () => {
    render(<Navbar />);

    expect(
      screen.getByRole("link", { name: "Code Self Study" })
    ).toHaveAttribute("href", "/");
    expect(screen.getByRole("link", { name: "About" })).toHaveAttribute(
      "href",
      "/about/"
    );
    expect(screen.getByRole("link", { name: "Events" })).toHaveAttribute(
      "href",
      "/events/"
    );
  });

  test("opens the mobile drawer when the menu button is clicked", async () => {
    const user = userEvent.setup();
    render(<Navbar />);

    // The drawer content is not mounted until it is opened.
    expect(screen.queryByText("Menu")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Open main menu" }));

    expect(await screen.findByText("Menu")).toBeInTheDocument();
  });

  test("closes the drawer when a menu link is clicked", async () => {
    const user = userEvent.setup();
    render(<Navbar />);

    await user.click(screen.getByRole("button", { name: "Open main menu" }));
    const dialog = await screen.findByRole("dialog");

    // The drawer lists the same links; clicking one dismisses the drawer.
    await user.click(within(dialog).getByRole("link", { name: "About" }));

    await waitFor(() =>
      expect(screen.queryByText("Menu")).not.toBeInTheDocument()
    );
  });

  test("does not duplicate the sign-in control inside the drawer", async () => {
    const user = userEvent.setup();
    render(<Navbar />);

    // Sign-in lives once in the top bar at every breakpoint; the drawer holds only
    // the nav links, so opening it must not mount a second SignInButton (a
    // duplicate would fire its own /api/me).
    await user.click(screen.getByRole("button", { name: "Open main menu" }));
    const dialog = await screen.findByRole("dialog");

    expect(
      within(dialog).queryByRole("button", { name: "Sign In" })
    ).not.toBeInTheDocument();
  });

  test("signed in: shows the username in the top bar, never the email", async () => {
    // Override the signed-out default from beforeEach with a signed-in /api/me.
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({
          email: "ada@example.com",
          username: "adalovelace",
        }),
      })
    );
    render(<Navbar />);

    expect(
      await screen.findByRole("button", {
        name: "Account menu for adalovelace",
      })
    ).toBeInTheDocument();
    expect(screen.getByText("adalovelace")).toBeInTheDocument();
    expect(screen.queryByText("ada@example.com")).not.toBeInTheDocument();
  });

  test("signed in: the username is there on the first frame, no Sign In flash", () => {
    // #367 end to end: with the server's flag cookie and a cached username, the
    // navbar must never paint the Sign In button while /api/me is in flight.
    document.cookie = `${AUTH_FLAG_COOKIE}=1`;
    writeAccountHint({ username: "adalovelace", avatar: "" });
    render(<Navbar />); // the beforeEach fetch stub is still pending here

    expect(screen.getByText("adalovelace")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Sign In" })
    ).not.toBeInTheDocument();
  });
});
