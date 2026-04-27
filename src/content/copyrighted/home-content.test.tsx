// @vitest-environment jsdom
import { afterEach, describe, expect, test } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";
import { HomeContent } from "./home-content";

afterEach(() => cleanup());

describe("HomeContent", () => {
  test("renders all four feature panes", () => {
    render(<HomeContent />);
    // One distinctive label / piece of content per pane.
    expect(screen.getByText("~/welcome")).toBeTruthy();
    expect(screen.getByText("~/languages")).toBeTruthy();
    expect(screen.getByText("~/stats")).toBeTruthy();
    expect(screen.getByText("recent activity · forum")).toBeTruthy();
  });

  test("uses the .terminal-home-grid layout class", () => {
    const { container } = render(<HomeContent />);
    expect(container.querySelector(".terminal-home-grid")).toBeTruthy();
  });
});
