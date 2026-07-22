import { beforeEach, describe, expect, test, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

type User = {
  id?: string;
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
    getAccessToken: vi.fn(async () => "tok_test"),
  },
}));

vi.mock("@workos-inc/authkit-react", () => ({
  useAuth: () => auth.value,
}));

// /api/me is exercised through a mocked apiFetch so the signed-in view can
// assert the server-returned email without a real request.
const apiFetchMock = vi.hoisted(() => vi.fn());
vi.mock("@/lib/api", () => ({ apiFetch: apiFetchMock }));

import SignInButton from "@/components/auth/SignInButton";

function meResolving(email: string) {
  return { ok: true, json: async () => ({ email }) };
}

describe("SignInButton", () => {
  beforeEach(() => {
    auth.value = {
      user: null,
      isLoading: false,
      signIn: vi.fn(),
      signOut: vi.fn(),
      getAccessToken: vi.fn(async () => "tok_test"),
    };
    apiFetchMock.mockReset();
    apiFetchMock.mockResolvedValue(meResolving("ada@example.com"));
  });

  test("signed out: shows a Sign In button and does not call the API", () => {
    render(<SignInButton />);

    expect(screen.getByRole("button", { name: "Sign In" })).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Sign Out" })
    ).not.toBeInTheDocument();
    expect(apiFetchMock).not.toHaveBeenCalled();
  });

  test("signed in: shows the email from /api/me and a Sign Out button", async () => {
    auth.value.user = { id: "u_ada", firstName: "Ada", lastName: "Lovelace" };
    render(<SignInButton />);

    expect(await screen.findByText("ada@example.com")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Sign Out" })
    ).toBeInTheDocument();
    expect(apiFetchMock).toHaveBeenCalledWith(
      "/api/me",
      auth.value.getAccessToken
    );
  });

  test("does not show the previous account's email after switching users", async () => {
    apiFetchMock
      .mockResolvedValueOnce(meResolving("ada@example.com"))
      .mockResolvedValueOnce(meResolving("grace@example.com"));

    auth.value.user = { id: "u_ada", firstName: "Ada" };
    const { rerender } = render(<SignInButton />);
    expect(await screen.findByText("ada@example.com")).toBeInTheDocument();

    auth.value.user = { id: "u_grace", firstName: "Grace" };
    rerender(<SignInButton />);

    // Bound to the identity: the instant we switch, the old email is gone even
    // before grace's /api/me resolves.
    expect(screen.queryByText("ada@example.com")).not.toBeInTheDocument();
    expect(await screen.findByText("grace@example.com")).toBeInTheDocument();
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
