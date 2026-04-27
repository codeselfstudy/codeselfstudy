// @vitest-environment jsdom
import { afterEach, describe, expect, test } from "vitest";
import { cleanup, render } from "@testing-library/react";
import { BlinkingCursor } from "./blinking-cursor";

afterEach(() => cleanup());

describe("BlinkingCursor", () => {
  test("renders ▮ glyph", () => {
    const { container } = render(<BlinkingCursor />);
    expect(container.textContent).toBe("▮");
  });

  test("has blink animation applied", () => {
    const { container } = render(<BlinkingCursor />);
    const span = container.querySelector("span");
    expect(span?.style.animation).toContain("blink");
  });
});
