import { describe, expect, test } from "vitest";
import { createMetadata } from "./metadata";

const TITLE = "Page";
const DESCRIPTION = "A description";

describe("createMetadata", () => {
  test("appends site name to title by default", () => {
    const meta = createMetadata({ title: TITLE, description: DESCRIPTION });
    expect(meta[0]).toEqual({ title: "Page | Code Self Study" });
  });

  test("uses bare title when isHome is true", () => {
    const meta = createMetadata({
      title: TITLE,
      description: DESCRIPTION,
      isHome: true,
    });
    expect(meta[0]).toEqual({ title: TITLE });
  });

  test("includes the description and og:description", () => {
    const meta = createMetadata({ title: TITLE, description: DESCRIPTION });
    expect(meta).toContainEqual({ name: "description", content: DESCRIPTION });
    expect(meta).toContainEqual({
      property: "og:description",
      content: DESCRIPTION,
    });
  });

  test("uses og title without site suffix even when not home", () => {
    const meta = createMetadata({ title: TITLE, description: DESCRIPTION });
    expect(meta).toContainEqual({ property: "og:title", content: TITLE });
  });

  test("sets robots to max-image-preview by default", () => {
    const meta = createMetadata({ title: TITLE, description: DESCRIPTION });
    expect(meta).toContainEqual({
      name: "robots",
      content: "max-image-preview:large",
    });
  });

  test("sets robots to noindex when isNoIndex is true", () => {
    const meta = createMetadata({
      title: TITLE,
      description: DESCRIPTION,
      isNoIndex: true,
    });
    expect(meta).toContainEqual({ name: "robots", content: "noindex" });
  });

  test("sets og:type to website", () => {
    const meta = createMetadata({ title: TITLE, description: DESCRIPTION });
    expect(meta).toContainEqual({ property: "og:type", content: "website" });
  });

  test("omits og:image entry when no image provided", () => {
    const meta = createMetadata({ title: TITLE, description: DESCRIPTION });
    expect(
      meta.some((entry) => "property" in entry && entry.property === "og:image")
    ).toBe(false);
  });

  test("appends og:image entry when image provided", () => {
    const image = "https://example.com/og.png";
    const meta = createMetadata({
      title: TITLE,
      description: DESCRIPTION,
      image,
    });
    expect(meta).toContainEqual({ property: "og:image", content: image });
  });
});
