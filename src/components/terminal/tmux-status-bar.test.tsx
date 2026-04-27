// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { act, cleanup, render, screen } from "@testing-library/react";
import { Clock, TmuxStatusBar } from "./tmux-status-bar";
import type { TmuxWindow } from "@/content/mock-home-data";

const WINDOWS: Array<TmuxWindow> = [
  { id: 0, label: "home" },
  { id: 1, label: "forum" },
  { id: 2, label: "events" },
];

afterEach(() => cleanup());

describe("TmuxStatusBar", () => {
  test("renders session label in brackets", () => {
    render(<TmuxStatusBar windows={WINDOWS} activeIndex={0} />);
    expect(screen.getByText("[codeselfstudy]")).toBeTruthy();
  });

  test("renders every window with id:label format", () => {
    render(
      <TmuxStatusBar
        windows={WINDOWS}
        activeIndex={1}
        session="codeselfstudy"
      />
    );
    expect(screen.getByText(/0:home/)).toBeTruthy();
    expect(screen.getByText(/1:forum/)).toBeTruthy();
    expect(screen.getByText(/2:events/)).toBeTruthy();
  });

  test("active window gets aria-current='page' and asterisk suffix", () => {
    render(<TmuxStatusBar windows={WINDOWS} activeIndex={1} />);
    const active = screen.getByText(/1:forum\*/);
    expect(active.getAttribute("aria-current")).toBe("page");
    const inactive = screen.getByText(/0:home/);
    expect(inactive.getAttribute("aria-current")).toBe(null);
  });

  test("renders custom 'right' slot", () => {
    render(
      <TmuxStatusBar
        windows={WINDOWS}
        activeIndex={0}
        right={<span>custom-right</span>}
      />
    );
    expect(screen.getByText("custom-right")).toBeTruthy();
  });
});

describe("Clock", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-04-26T18:42:00"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  test("renders HH:MM after mount and updates after 30s", () => {
    render(<Clock />);
    act(() => {
      vi.advanceTimersByTime(0);
    });
    expect(screen.getByLabelText("current time").textContent).toBe("18:42");

    vi.setSystemTime(new Date("2026-04-26T18:43:00"));
    act(() => {
      vi.advanceTimersByTime(30_000);
    });
    expect(screen.getByLabelText("current time").textContent).toBe("18:43");
  });
});
