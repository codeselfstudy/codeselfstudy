// @vitest-environment jsdom
import { afterEach, describe, expect, test } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";
import { TermPane } from "./term-pane";

afterEach(() => cleanup());

describe("TermPane", () => {
  test("renders index, label, and child body", () => {
    render(
      <TermPane index={2} label="~/welcome">
        <div>body</div>
      </TermPane>
    );
    expect(screen.getByText("2")).toBeTruthy();
    expect(screen.getByText("~/welcome")).toBeTruthy();
    expect(screen.getByText("body")).toBeTruthy();
  });

  test("renders keys hint when provided", () => {
    render(
      <TermPane index={0} label="~/x" keys="↑↓ scroll">
        <span />
      </TermPane>
    );
    expect(screen.getByText("↑↓ scroll")).toBeTruthy();
  });

  test("variant='full' applies grid-column 1 / -1", () => {
    const { container } = render(
      <TermPane index={0} label="~/x" variant="full">
        <span />
      </TermPane>
    );
    const section = container.querySelector("section");
    expect(section?.style.gridColumn).toBe("1 / -1");
  });

  test("variant='last' drops the right border", () => {
    const { container } = render(
      <TermPane index={0} label="~/x" variant="last">
        <span />
      </TermPane>
    );
    const section = container.querySelector("section");
    expect(section?.style.borderRightWidth).toBe("0px");
  });

  test("labelColor='magenta' colors the label with --term-magenta", () => {
    const { container } = render(
      <TermPane index={3} label="recent" labelColor="magenta">
        <span />
      </TermPane>
    );
    const labelSpan = container.querySelector("header span:nth-child(2)");
    expect((labelSpan as HTMLElement | null)?.style.color).toContain(
      "--term-magenta"
    );
  });
});
