// @vitest-environment jsdom
import { afterEach, describe, expect, test, vi } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";

import { PageWrapper } from "./page-wrapper";

vi.mock("@tanstack/react-router", () => ({
  useLocation: () => ({ pathname: "/about" }),
}));

afterEach(() => cleanup());

describe("PageWrapper", () => {
  test("wraps children in a TermPane", () => {
    render(
      <PageWrapper>
        <p>hello</p>
      </PageWrapper>
    );
    expect(screen.getByText("hello")).toBeTruthy();
    expect(screen.getByText("~/about")).toBeTruthy();
  });

  test("custom paneTitle overrides the derived label", () => {
    render(
      <PageWrapper paneTitle="~/custom">
        <span>x</span>
      </PageWrapper>
    );
    expect(screen.getByText("~/custom")).toBeTruthy();
    expect(screen.queryByText("~/about")).toBeNull();
  });

  test("paneIndex is rendered in the pane bar", () => {
    render(
      <PageWrapper paneIndex={5} paneTitle="~/x">
        <span />
      </PageWrapper>
    );
    expect(screen.getByText("5")).toBeTruthy();
  });
});
