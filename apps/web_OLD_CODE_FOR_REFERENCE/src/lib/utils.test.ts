import { describe, expect, test } from "vitest";
import { cn } from "./utils";

describe("cn", () => {
  test("joins plain string class names", () => {
    expect(cn("a", "b", "c")).toBe("a b c");
  });

  test("filters out falsy values", () => {
    expect(cn("a", false, null, undefined, "", 0, "b")).toBe("a b");
  });

  test("flattens nested arrays", () => {
    expect(cn(["a", ["b", ["c"]]])).toBe("a b c");
  });

  test("supports conditional class objects", () => {
    expect(cn({ a: true, b: false, c: true })).toBe("a c");
  });

  test("returns an empty string when given no inputs", () => {
    expect(cn()).toBe("");
  });

  test("dedupes conflicting tailwind utilities, last-wins", () => {
    expect(cn("p-2", "p-4")).toBe("p-4");
    expect(cn("text-sm", "text-lg")).toBe("text-lg");
  });

  test("preserves non-conflicting tailwind utilities", () => {
    expect(cn("p-2", "m-4", "text-red-500")).toBe("p-2 m-4 text-red-500");
  });

  test("merges responsive variants independently of base utilities", () => {
    expect(cn("p-2", "md:p-4")).toBe("p-2 md:p-4");
  });
});
