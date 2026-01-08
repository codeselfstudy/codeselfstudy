/**
 * Tests for the redirect system.
 */
import { describe, expect, test } from "vitest";
import { findRedirect } from "./redirects";

describe("findRedirect", () => {
  test("should find exact match redirects", () => {
    const result = findRedirect("/book");
    expect(result).toEqual({ to: "/learn", status: 308 });
  });

  test("should find exact match with trailing slash", () => {
    const result = findRedirect("/book/");
    expect(result).toEqual({ to: "/learn", status: 308 });
  });

  test("should return null for non-existent paths", () => {
    const result = findRedirect("/nonexistent");
    expect(result).toBeNull();
  });

  test("should match wildcard patterns - /blog/*", () => {
    const result = findRedirect("/blog/some-article");
    expect(result).toEqual({ to: "/learn", status: 308 });
  });

  test("should match wildcard patterns - /blog/* with nested paths", () => {
    const result = findRedirect("/blog/2024/01/article");
    expect(result).toEqual({ to: "/learn", status: 308 });
  });

  test("should match wildcard patterns - /wiki/*", () => {
    const result = findRedirect("/wiki/some-page");
    expect(result).toEqual({ to: "/learn", status: 308 });
  });

  test("should match wildcard patterns - /b/*", () => {
    const result = findRedirect("/b/some-post");
    expect(result).toEqual({ to: "/learn", status: 308 });
  });

  test("should prefer exact matches over wildcard patterns", () => {
    // /wiki and /wiki/ are exact matches to /learn
    const resultExact = findRedirect("/wiki");
    expect(resultExact).toEqual({ to: "/learn", status: 308 });

    // /wiki/Main_Page is exact match to /contact
    const resultMainPage = findRedirect("/wiki/Main_Page");
    expect(resultMainPage).toEqual({ to: "/contact", status: 308 });

    // /wiki/something-else should match wildcard to /learn
    const resultWildcard = findRedirect("/wiki/something-else");
    expect(resultWildcard).toEqual({ to: "/learn", status: 308 });
  });

  test("should not match wildcard without trailing slash", () => {
    // /b itself is an exact match
    const result = findRedirect("/b");
    expect(result).toEqual({ to: "/learn", status: 308 });
  });

  test("should not match partial prefixes", () => {
    // /blogpost should not match /blog/*
    const result = findRedirect("/blogpost");
    expect(result).toBeNull();
  });

  test("should handle trailing slash in wildcard patterns", () => {
    // /wiki/ is an exact match, not a wildcard match
    const result = findRedirect("/wiki/");
    expect(result).toEqual({ to: "/learn", status: 308 });
  });

  test("should not match path with only trailing slash for wildcards", () => {
    // /blog/ should not match /blog/* (no content after slash)
    const result = findRedirect("/blog/");
    expect(result).toBeNull();
  });
});
