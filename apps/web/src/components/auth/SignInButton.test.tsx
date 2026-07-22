import { beforeEach, describe, expect, test, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

type User = {
  firstName?: string | null;
  lastName?: string | null;
  profilePictureUrl?: string | null;
};

// Drive SignInButton through a mocked useAuth so both auth states render without
// a real WorkOS session.
const auth = vi.hoisted(() => ({
  value: {
    user: null as User | null,
    isLoading: false,
    signIn: vi.fn(),
    signOut: vi.fn(),
  },
}));

vi.mock("@workos-inc/authkit-react", () => ({
  useAuth: () => auth.value,
}));

import SignInButton from "@/components/auth/SignInButton";

describe("SignInButton", () => {
  beforeEach(() => {
    auth.value = {
      user: null,
      isLoading: false,
      signIn: vi.fn(),
      signOut: vi.fn(),
    };
  });

  test("signed out: shows a Sign In button", () => {
    render(<SignInButton />);

    expect(screen.getByRole("button", { name: "Sign In" })).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Sign Out" })
    ).not.toBeInTheDocument();
  });

  test("signed in: shows the name and a Sign Out button", () => {
    auth.value.user = { firstName: "Ada", lastName: "Lovelace" };
    render(<SignInButton />);

    expect(screen.getByText("Ada Lovelace")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Sign Out" })
    ).toBeInTheDocument();
  });

  test("clicking Sign In launches the WorkOS flow with the current path as returnTo", async () => {
    const user = userEvent.setup();
    render(<SignInButton />);

    await user.click(screen.getByRole("button", { name: "Sign In" }));

    expect(auth.value.signIn).toHaveBeenCalledWith({
      state: { returnTo: window.location.pathname },
    });
  });
});
