// @vitest-environment jsdom
import { afterEach, describe, expect, test, vi } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";

import { NAV_LINKS, TermNav } from "./term-nav";

vi.mock("@tanstack/react-router", () => ({
  Link: ({
    to,
    children,
    ...rest
  }: {
    to: string;
    children: React.ReactNode;
  }) => (
    <a href={to} {...rest}>
      {children}
    </a>
  ),
}));

afterEach(() => cleanup());

describe("TermNav", () => {
  test("renders the brand", () => {
    render(<TermNav />);
    expect(screen.getByText("codeselfstudy")).toBeTruthy();
  });

  test("renders every NAV_LINKS entry as an <a> with correct href", () => {
    render(<TermNav />);
    for (const link of NAV_LINKS) {
      const a = screen.getByRole("link", { name: link.label });
      expect(a.getAttribute("href")).toBe(link.to);
    }
  });

  test("primary nav landmark has aria-label='primary'", () => {
    render(<TermNav />);
    const nav = screen.getByRole("navigation", { name: "primary" });
    expect(nav).toBeTruthy();
  });

  test("themeToggle slot renders when provided", () => {
    render(<TermNav themeToggle={<button>tt</button>} />);
    expect(screen.getByRole("button", { name: "tt" })).toBeTruthy();
  });

  test("custom links prop overrides the default NAV_LINKS", () => {
    render(
      <TermNav
        links={[
          { to: "/x", label: "x-page" },
          { to: "/y", label: "y-page" },
        ]}
      />
    );
    expect(screen.getAllByRole("link")).toHaveLength(2);
    expect(screen.getByRole("link", { name: "x-page" })).toBeTruthy();
  });
});
