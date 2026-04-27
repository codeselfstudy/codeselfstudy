// @vitest-environment jsdom
import { afterEach, describe, expect, test } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";
import { LanguagesPane } from "./languages-pane";
import { LANGUAGES } from "@/content/mock-home-data";

afterEach(() => cleanup());

describe("LanguagesPane", () => {
  test("renders one row per language plus the overflow row", () => {
    render(<LanguagesPane />);
    for (const lang of LANGUAGES) {
      expect(screen.getByText(lang.name)).toBeTruthy();
      expect(screen.getByText(`${lang.percent}%`)).toBeTruthy();
    }
    expect(screen.getByText(/\+ 12 more/)).toBeTruthy();
  });

  test("first row is python at 62%", () => {
    render(<LanguagesPane />);
    expect(screen.getByText("python")).toBeTruthy();
    expect(screen.getByText("62%")).toBeTruthy();
  });

  test("custom languages prop overrides default", () => {
    render(
      <LanguagesPane
        languages={[{ name: "lisp", percent: 100, color: "magenta" }]}
        overflowLabel=""
      />
    );
    expect(screen.getByText("lisp")).toBeTruthy();
    expect(screen.getByText("100%")).toBeTruthy();
    expect(screen.queryByText("python")).toBeNull();
  });
});
