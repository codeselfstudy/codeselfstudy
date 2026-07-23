import { afterEach, describe, expect, test, vi } from "vitest";
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import ThemeToggle from "@/components/ThemeToggle";

// `test/setup.ts` only installs its `matchMedia` stub when jsdom has none, and
// jsdom ships one — so the component gets jsdom's real MediaQueryList, which
// never fires a "change" event. Swap in a controllable one whose listeners we
// can invoke by hand to drive the OS-preference path.
function stubMatchMedia() {
  const listeners = new Set<(event: MediaQueryListEvent) => void>();

  vi.stubGlobal(
    "matchMedia",
    vi.fn(() => ({
      matches: false,
      media: "(prefers-color-scheme: dark)",
      addEventListener: (
        _type: string,
        listener: (event: MediaQueryListEvent) => void
      ) => listeners.add(listener),
      removeEventListener: (
        _type: string,
        listener: (event: MediaQueryListEvent) => void
      ) => listeners.delete(listener),
    }))
  );

  // The listener calls setState outside React's event system, so wrap it in act.
  return (matches: boolean) =>
    act(async () => {
      for (const listener of listeners) {
        listener({ matches } as MediaQueryListEvent);
      }
    });
}

function getToggle() {
  return screen.getByRole("button", { name: "Toggle dark mode" });
}

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
  // The component mutates the real <html> element and localStorage, so both
  // leak into the next test unless they are reset here.
  document.documentElement.className = "";
  localStorage.clear();
});

describe("ThemeToggle", () => {
  test("clicking from light turns dark mode on and stores the choice", async () => {
    const user = userEvent.setup();
    render(<ThemeToggle />);

    expect(getToggle()).toHaveAttribute("aria-pressed", "false");

    await user.click(getToggle());

    expect(document.documentElement).toHaveClass("dark");
    expect(localStorage.getItem("theme")).toBe("dark");
    expect(getToggle()).toHaveAttribute("aria-pressed", "true");
  });

  test("mounting with dark already set reflects it, and clicking turns it off", async () => {
    // The inline script in Layout.astro sets this class before the island
    // hydrates, so the initial state has to be read off the DOM, not assumed.
    document.documentElement.classList.add("dark");
    const user = userEvent.setup();
    render(<ThemeToggle />);

    expect(getToggle()).toHaveAttribute("aria-pressed", "true");

    await user.click(getToggle());

    expect(document.documentElement).not.toHaveClass("dark");
    expect(localStorage.getItem("theme")).toBe("light");
    expect(getToggle()).toHaveAttribute("aria-pressed", "false");
  });

  test("follows the OS setting while the user has not chosen explicitly", async () => {
    const fireChange = stubMatchMedia();
    render(<ThemeToggle />);

    await fireChange(true);

    expect(document.documentElement).toHaveClass("dark");
    expect(getToggle()).toHaveAttribute("aria-pressed", "true");
  });

  test("ignores the OS setting once a theme is stored", async () => {
    localStorage.setItem("theme", "light");
    const fireChange = stubMatchMedia();
    render(<ThemeToggle />);

    await fireChange(true);

    expect(document.documentElement).not.toHaveClass("dark");
    expect(getToggle()).toHaveAttribute("aria-pressed", "false");
  });

  test("treats an unreadable localStorage as no stored choice", async () => {
    // Safari in private mode and cookie-blocking extensions can make these
    // throw; the component swallows that rather than breaking the navbar.
    vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new Error("storage blocked");
    });
    const fireChange = stubMatchMedia();
    render(<ThemeToggle />);

    await fireChange(true);

    expect(document.documentElement).toHaveClass("dark");
  });

  test("still toggles when the choice cannot be persisted", async () => {
    vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new Error("storage blocked");
    });
    const user = userEvent.setup();
    render(<ThemeToggle />);

    await user.click(getToggle());

    expect(document.documentElement).toHaveClass("dark");
    expect(getToggle()).toHaveAttribute("aria-pressed", "true");
  });

  test("works in a browser without matchMedia", async () => {
    vi.stubGlobal("matchMedia", undefined);
    const user = userEvent.setup();
    render(<ThemeToggle />);

    await user.click(getToggle());

    expect(document.documentElement).toHaveClass("dark");
  });
});
