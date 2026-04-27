// @vitest-environment jsdom
import { afterEach, describe, expect, test } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";
import { ForumActivityPane } from "./forum-activity-pane";
import { FORUM_ACTIVITY } from "@/content/mock-home-data";

afterEach(() => cleanup());

describe("ForumActivityPane", () => {
  test("renders one row per post in tbody", () => {
    render(<ForumActivityPane />);
    const table = screen.getByRole("table");
    const tbody = table.querySelector("tbody");
    const rows = tbody?.querySelectorAll("tr") ?? [];
    expect(rows).toHaveLength(FORUM_ACTIVITY.length);
  });

  test("first row contains alice / #rust / 8 replies", () => {
    render(<ForumActivityPane />);
    expect(screen.getByText("alice")).toBeTruthy();
    expect(screen.getByText("#rust")).toBeTruthy();
    expect(screen.getByText("8")).toBeTruthy();
    expect(screen.getByText(/code review on a tiny json parser/)).toBeTruthy();
  });

  test("table has the expected column headers", () => {
    render(<ForumActivityPane />);
    for (const header of ["when", "by", "category", "topic", "replies"]) {
      expect(screen.getByRole("columnheader", { name: header })).toBeTruthy();
    }
  });

  test("custom posts prop overrides the default", () => {
    render(
      <ForumActivityPane
        posts={[
          {
            when: "1y",
            by: "zelda",
            category: "#test",
            categoryColor: "#fff",
            topic: "test post",
            replies: 99,
          },
        ]}
      />
    );
    expect(screen.getByText("zelda")).toBeTruthy();
    expect(screen.getByText("test post")).toBeTruthy();
    expect(screen.queryByText("alice")).toBeNull();
  });
});
