import { afterEach, describe, expect, test, vi } from "vitest";

import {
  AUTH_FLAG_COOKIE,
  clearAccountHint,
  isSignedInHint,
  readAccountHint,
  writeAccountHint,
} from "./authHint";

// jsdom gives real document.cookie and localStorage, so these exercise the same
// code paths the browser will. Both are process-wide, so reset after each test.
function setCookie(raw: string) {
  document.cookie = raw;
}

afterEach(() => {
  vi.restoreAllMocks();
  localStorage.clear();
  for (const part of document.cookie.split(";")) {
    const name = part.split("=")[0]?.trim();
    if (name) document.cookie = `${name}=; max-age=0`;
  }
});

describe("isSignedInHint", () => {
  test("false when the flag cookie is absent", () => {
    expect(isSignedInHint()).toBe(false);
  });

  test("true when the server has set the flag", () => {
    setCookie(`${AUTH_FLAG_COOKIE}=1`);
    expect(isSignedInHint()).toBe(true);
  });

  test("finds the flag among other cookies", () => {
    setCookie("theme=dark");
    setCookie(`${AUTH_FLAG_COOKIE}=1`);
    setCookie("other=x");
    expect(isSignedInHint()).toBe(true);
  });

  test("a cookie whose name merely ends with the flag name does not count", () => {
    // Substring matching here would let an unrelated cookie forge a signed-in
    // first paint.
    setCookie(`x${AUTH_FLAG_COOKIE}=1`);
    expect(isSignedInHint()).toBe(false);
  });
});

describe("readAccountHint", () => {
  test("null when nothing has been stored", () => {
    expect(readAccountHint()).toBeNull();
  });

  test("round-trips what writeAccountHint stored", () => {
    writeAccountHint({ username: "adalovelace", avatar: "https://i/a.png" });
    expect(readAccountHint()).toEqual({
      username: "adalovelace",
      avatar: "https://i/a.png",
    });
  });

  test("null when the stored value is not JSON", () => {
    localStorage.setItem("auth:account", "{not json");
    expect(readAccountHint()).toBeNull();
  });

  test("null when the stored JSON is not an object", () => {
    localStorage.setItem("auth:account", '"adalovelace"');
    expect(readAccountHint()).toBeNull();
  });

  test("null when the stored JSON is literal null", () => {
    localStorage.setItem("auth:account", "null");
    expect(readAccountHint()).toBeNull();
  });

  test("null when the fields are the wrong shape", () => {
    // An older release, another tab, or someone with devtools open.
    localStorage.setItem("auth:account", '{"username":42,"avatar":"x"}');
    expect(readAccountHint()).toBeNull();
    localStorage.setItem("auth:account", '{"username":"ada"}');
    expect(readAccountHint()).toBeNull();
  });

  test("null when storage itself throws", () => {
    vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new Error("storage disabled");
    });
    expect(readAccountHint()).toBeNull();
  });
});

describe("writeAccountHint", () => {
  test("swallows a storage failure — the hint is an optimization", () => {
    vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new Error("quota exceeded");
    });
    expect(() =>
      writeAccountHint({ username: "ada", avatar: "" })
    ).not.toThrow();
  });
});

describe("clearAccountHint", () => {
  test("removes a stored hint", () => {
    writeAccountHint({ username: "ada", avatar: "" });
    clearAccountHint();
    expect(readAccountHint()).toBeNull();
  });

  test("swallows a storage failure", () => {
    vi.spyOn(Storage.prototype, "removeItem").mockImplementation(() => {
      throw new Error("storage disabled");
    });
    expect(() => clearAccountHint()).not.toThrow();
  });
});
