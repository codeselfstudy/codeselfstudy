import { describe, expect, test, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";

// AuthKit-react initializes a real WorkOS browser session on mount; stub the
// provider so this stays a pure render test with no network.
vi.mock("@workos-inc/authkit-react", () => ({
  AuthKitProvider: ({ children }: { children: ReactNode }) => <>{children}</>,
}));

import AuthProvider from "@/integrations/workos/AuthProvider";

describe("AuthProvider", () => {
  test("renders its children inside the AuthKit provider", () => {
    render(
      <AuthProvider>
        <span>protected child</span>
      </AuthProvider>
    );

    expect(screen.getByText("protected child")).toBeInTheDocument();
  });
});
