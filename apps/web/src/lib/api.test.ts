import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useApiFetch } from "./api";

const { useAuthMock } = vi.hoisted(() => ({ useAuthMock: vi.fn() }));

vi.mock("@workos-inc/authkit-react", () => ({
  useAuth: () => useAuthMock(),
}));

describe("useApiFetch", () => {
  let originalFetch: typeof fetch;

  beforeEach(() => {
    originalFetch = global.fetch;
    global.fetch = vi
      .fn()
      .mockResolvedValue(new Response(null, { status: 204 }));
  });

  afterEach(() => {
    global.fetch = originalFetch;
    useAuthMock.mockReset();
  });

  it("attaches the bearer token to /api requests", async () => {
    useAuthMock.mockReturnValue({
      getAccessToken: vi.fn().mockResolvedValue("token-123"),
    });

    const { result } = renderHook(() => useApiFetch());
    await act(async () => {
      await result.current("/api/me");
    });

    expect(global.fetch).toHaveBeenCalledTimes(1);
    const [, init] = (global.fetch as ReturnType<typeof vi.fn>).mock.calls[0];
    const headers = new Headers(init.headers);
    expect(headers.get("Authorization")).toBe("Bearer token-123");
    expect(headers.get("Accept")).toBe("application/json");
  });

  it("omits Authorization when getAccessToken returns null", async () => {
    useAuthMock.mockReturnValue({
      getAccessToken: vi.fn().mockResolvedValue(null),
    });

    const { result } = renderHook(() => useApiFetch());
    await act(async () => {
      await result.current("/api/me");
    });

    const [, init] = (global.fetch as ReturnType<typeof vi.fn>).mock.calls[0];
    const headers = new Headers(init.headers);
    expect(headers.has("Authorization")).toBe(false);
  });

  it("preserves caller-supplied headers and method", async () => {
    useAuthMock.mockReturnValue({
      getAccessToken: vi.fn().mockResolvedValue("t"),
    });

    const { result } = renderHook(() => useApiFetch());
    await act(async () => {
      await result.current("/api/todos", {
        method: "POST",
        headers: { "X-Custom": "yes", Accept: "text/plain" },
        body: JSON.stringify({ title: "buy milk" }),
      });
    });

    const [, init] = (global.fetch as ReturnType<typeof vi.fn>).mock.calls[0];
    const headers = new Headers(init.headers);
    expect(init.method).toBe("POST");
    expect(headers.get("X-Custom")).toBe("yes");
    // Caller-supplied Accept wins over the default we'd add.
    expect(headers.get("Accept")).toBe("text/plain");
    expect(headers.get("Authorization")).toBe("Bearer t");
  });
});
