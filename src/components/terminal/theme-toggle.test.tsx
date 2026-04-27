// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, test } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { ThemeToggle } from "./theme-toggle";

beforeEach(() => {
  document.documentElement.removeAttribute("data-theme");
  window.localStorage.clear();
});

afterEach(() => {
  cleanup();
});

describe("ThemeToggle", () => {
  test("renders button with default 'dark' label", () => {
    render(<ThemeToggle />);
    const btn = screen.getByRole("button", { name: /toggle light and dark/i });
    expect(btn.textContent).toContain("dark");
  });

  test("click flips data-theme to light and persists", () => {
    render(<ThemeToggle />);
    const btn = screen.getByRole("button", { name: /toggle light and dark/i });

    fireEvent.click(btn);
    expect(document.documentElement.getAttribute("data-theme")).toBe("light");
    expect(window.localStorage.getItem("tmux-theme")).toBe("light");
    expect(btn.textContent).toContain("light");

    fireEvent.click(btn);
    expect(document.documentElement.getAttribute("data-theme")).toBe(null);
    expect(window.localStorage.getItem("tmux-theme")).toBe("dark");
    expect(btn.textContent).toContain("dark");
  });

  test("Enter and Space key activate the toggle", () => {
    render(<ThemeToggle />);
    const btn = screen.getByRole("button", { name: /toggle light and dark/i });

    fireEvent.keyDown(btn, { key: "Enter" });
    expect(document.documentElement.getAttribute("data-theme")).toBe("light");

    fireEvent.keyDown(btn, { key: " " });
    expect(document.documentElement.getAttribute("data-theme")).toBe(null);
  });

  test("initial render reads stored theme from localStorage", () => {
    window.localStorage.setItem("tmux-theme", "light");
    render(<ThemeToggle />);
    expect(document.documentElement.getAttribute("data-theme")).toBe("light");
    const btn = screen.getByRole("button", { name: /toggle light and dark/i });
    expect(btn.textContent).toContain("light");
  });
});
