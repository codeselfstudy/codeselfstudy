// @vitest-environment jsdom
import { afterEach, describe, expect, test } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";
import { Meter } from "./meter";

afterEach(() => cleanup());

describe("Meter", () => {
  test("inner fill width matches the value prop", () => {
    const { container } = render(<Meter value={73} />);
    const inner = container.querySelector("i");
    expect(inner?.style.width).toBe("73%");
  });

  test("clamps values above 100", () => {
    const { container } = render(<Meter value={250} />);
    const inner = container.querySelector("i");
    expect(inner?.style.width).toBe("100%");
  });

  test("exposes aria meter semantics", () => {
    render(<Meter value={42} ariaLabel="members" />);
    const meter = screen.getByRole("meter", { name: "members" });
    expect(meter.getAttribute("aria-valuenow")).toBe("42");
    expect(meter.getAttribute("aria-valuemin")).toBe("0");
    expect(meter.getAttribute("aria-valuemax")).toBe("100");
  });

  test("color variant maps to --term-magenta token", () => {
    const { container } = render(<Meter value={50} color="magenta" />);
    const inner = container.querySelector("i");
    expect(inner?.style.background).toContain("--term-magenta");
  });
});
