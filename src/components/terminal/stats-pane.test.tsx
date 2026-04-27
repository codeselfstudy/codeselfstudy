// @vitest-environment jsdom
import { afterEach, describe, expect, test } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";
import { StatsPane } from "./stats-pane";
import { STATS } from "@/content/mock-home-data";

afterEach(() => cleanup());

describe("StatsPane", () => {
  test("renders all three community meters", () => {
    render(<StatsPane />);
    for (const row of STATS.community) {
      expect(screen.getByText(row.label)).toBeTruthy();
      expect(screen.getByText(String(row.value))).toBeTruthy();
      expect(screen.getByRole("meter", { name: row.label })).toBeTruthy();
    }
  });

  test("renders the net-this-week lines", () => {
    render(<StatsPane />);
    for (const row of STATS.netThisWeek) {
      expect(screen.getByText(row.label)).toBeTruthy();
      expect(screen.getByText(String(row.value))).toBeTruthy();
    }
  });

  test("renders the load-avg ASCII string and caption", () => {
    const { container } = render(<StatsPane />);
    expect(container.textContent).toContain(STATS.loadAvgAscii);
    expect(container.textContent).toContain(STATS.loadAvgCaption);
  });

  test("custom stats prop overrides default", () => {
    render(
      <StatsPane
        stats={{
          ...STATS,
          community: [
            {
              label: "custom-label",
              value: "999",
              color: "green",
              meterPercent: 50,
            },
          ],
        }}
      />
    );
    expect(screen.getByText("custom-label")).toBeTruthy();
    expect(screen.getByText("999")).toBeTruthy();
    expect(screen.queryByText("members")).toBeNull();
  });
});
