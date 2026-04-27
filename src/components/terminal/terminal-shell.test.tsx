// @vitest-environment jsdom
import { afterEach, describe, expect, test } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";
import { TerminalShell } from "./terminal-shell";

afterEach(() => cleanup());

describe("TerminalShell", () => {
  test("renders nav, children, and statusBar in DOM order", () => {
    const { container } = render(
      <TerminalShell
        nav={<nav data-testid="nav">NAV</nav>}
        statusBar={<div data-testid="status">STATUS</div>}
      >
        <p data-testid="body">BODY</p>
      </TerminalShell>
    );
    const order = Array.from(container.querySelectorAll("[data-testid]")).map(
      (el) => el.getAttribute("data-testid")
    );
    expect(order).toEqual(["nav", "body", "status"]);
  });

  test("applies the .terminal-shell class for the global token-driven look", () => {
    const { container } = render(
      <TerminalShell nav={null} statusBar={null}>
        <span />
      </TerminalShell>
    );
    expect(
      container.firstElementChild?.classList.contains("terminal-shell")
    ).toBe(true);
  });

  test("body is wrapped in a <main> landmark", () => {
    render(
      <TerminalShell nav={<header />} statusBar={<footer />}>
        <p>page</p>
      </TerminalShell>
    );
    const main = screen.getByRole("main");
    expect(main.textContent).toBe("page");
  });
});
