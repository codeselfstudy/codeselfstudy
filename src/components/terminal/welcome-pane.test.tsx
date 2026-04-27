// @vitest-environment jsdom
import { afterEach, describe, expect, test } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";
import { WelcomePane } from "./welcome-pane";
import { WELCOME_COPY } from "@/content/mock-home-data";

afterEach(() => cleanup());

describe("WelcomePane", () => {
  test("renders title text", () => {
    render(<WelcomePane />);
    expect(screen.getByText(WELCOME_COPY.title)).toBeTruthy();
  });

  test("renders every whatWeDo bullet", () => {
    render(<WelcomePane />);
    for (const item of WELCOME_COPY.whatWeDo) {
      expect(screen.getByText(item)).toBeTruthy();
    }
  });

  test("CTA renders with correct label and href", () => {
    render(<WelcomePane />);
    const cta = screen.getByRole("link", { name: WELCOME_COPY.cta.label });
    expect(cta.getAttribute("href")).toBe(WELCOME_COPY.cta.href);
  });

  test("highlights the leadHighlight substring", () => {
    render(<WelcomePane />);
    const bold = screen.getByText(WELCOME_COPY.leadHighlight);
    expect(bold.tagName.toLowerCase()).toBe("b");
  });

  test("renders the blinking cursor glyph", () => {
    const { container } = render(<WelcomePane />);
    expect(container.textContent).toContain("▮");
  });

  test("custom copy prop overrides default", () => {
    render(
      <WelcomePane
        copy={{
          ...WELCOME_COPY,
          title: "different title",
          cta: { label: "Sign up", href: "https://example.com" },
        }}
      />
    );
    expect(screen.getByText("different title")).toBeTruthy();
    expect(
      screen.getByRole("link", { name: "Sign up" }).getAttribute("href")
    ).toBe("https://example.com");
  });
});
