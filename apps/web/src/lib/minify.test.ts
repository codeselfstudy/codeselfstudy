import { describe, expect, test } from "vitest";
import { minify } from "./minify";

describe("minify", () => {
  test("flattens an indented multi-line string to one line", () => {
    const input = `
      a
        b
      c
    `;
    expect(minify(input)).toBe("a b c");
  });

  test("trims leading and trailing whitespace on each line", () => {
    expect(minify("  a  \n\tb\t")).toBe("a b");
  });

  test("drops blank and whitespace-only lines", () => {
    expect(minify("a\n\n   \nb")).toBe("a b");
  });

  test("handles CRLF line endings (trim strips the trailing \\r)", () => {
    expect(minify("a\r\nb")).toBe("a b");
  });

  test("leaves a single line unchanged (aside from trimming)", () => {
    expect(minify("  one line  ")).toBe("one line");
  });

  test("returns an empty string for empty or whitespace-only input", () => {
    expect(minify("")).toBe("");
    expect(minify("\n  \n\t\n")).toBe("");
  });

  test("flattens the Google Analytics snippet to one line", () => {
    const snippet = `
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());

  gtag('config', 'G-99DC6WSRWL');
`;
    expect(minify(snippet)).toBe(
      "window.dataLayer = window.dataLayer || []; " +
        "function gtag(){dataLayer.push(arguments);} " +
        "gtag('js', new Date()); " +
        "gtag('config', 'G-99DC6WSRWL');"
    );
  });
});
