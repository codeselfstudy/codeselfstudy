// @vitest-environment jsdom
import { afterEach, describe, expect, test } from "vitest";
import { cleanup, render } from "@testing-library/react";
import { AsciiBar } from "./ascii-bar";

afterEach(() => cleanup());

describe("AsciiBar", () => {
  test("renders proportional fill at 50%", () => {
    const { container } = render(<AsciiBar value={50} width={10} />);
    expect(container.textContent).toBe("█████░░░░░");
  });

  test("clamps values above 100", () => {
    const { container } = render(<AsciiBar value={250} width={4} />);
    expect(container.textContent).toBe("████");
  });

  test("clamps negative values to 0", () => {
    const { container } = render(<AsciiBar value={-5} width={4} />);
    expect(container.textContent).toBe("░░░░");
  });

  test("color variant maps to --term-cyan token", () => {
    const { container } = render(
      <AsciiBar value={20} width={5} color="cyan" />
    );
    const span = container.querySelector("span");
    expect(span?.style.color).toContain("--term-cyan");
  });
});
