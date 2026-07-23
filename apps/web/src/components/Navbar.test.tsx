import { describe, expect, test, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";

// Navbar wraps its content in the WorkOS AuthProvider and renders SignInButton;
// stub AuthKit so the nav tests don't spin up a real browser session.
vi.mock("@workos-inc/authkit-react", () => ({
  AuthKitProvider: ({ children }: { children: ReactNode }) => children,
  useAuth: () => ({
    user: null,
    isLoading: false,
    signIn: vi.fn(),
    signOut: vi.fn(),
  }),
}));

import Navbar from "@/components/Navbar";

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
});
