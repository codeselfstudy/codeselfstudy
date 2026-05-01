import { act, renderHook } from "@testing-library/react";
import {
  afterEach,
  beforeEach,
  describe,
  expect,
  test,
  vi,
} from "vitest";
import { useIsMobile } from "./use-mobile";

const MOBILE_BREAKPOINT = 768;

type ChangeListener = (event: MediaQueryListEvent) => void;

interface FakeMediaQueryList {
  matches: boolean;
  media: string;
  addEventListener: ReturnType<typeof vi.fn>;
  removeEventListener: ReturnType<typeof vi.fn>;
  __fire: (matches: boolean) => void;
}

function installMatchMedia(): { mql: FakeMediaQueryList; restore: () => void } {
  let listener: ChangeListener | null = null;
  const mql: FakeMediaQueryList = {
    matches: false,
    media: "",
    addEventListener: vi.fn((_event: string, handler: ChangeListener) => {
      listener = handler;
    }),
    removeEventListener: vi.fn(() => {
      listener = null;
    }),
    __fire: (matches: boolean) => {
      mql.matches = matches;
      listener?.({ matches } as MediaQueryListEvent);
    },
  };
  const matchMedia = vi.fn((query: string) => {
    mql.media = query;
    return mql as unknown as MediaQueryList;
  });
  const original = window.matchMedia;
  Object.defineProperty(window, "matchMedia", {
    configurable: true,
    writable: true,
    value: matchMedia,
  });
  return {
    mql,
    restore: () => {
      Object.defineProperty(window, "matchMedia", {
        configurable: true,
        writable: true,
        value: original,
      });
    },
  };
}

function setInnerWidth(width: number) {
  Object.defineProperty(window, "innerWidth", {
    configurable: true,
    writable: true,
    value: width,
  });
}

describe("useIsMobile", () => {
  let restoreMatchMedia: () => void;
  let mql: FakeMediaQueryList;

  beforeEach(() => {
    const installed = installMatchMedia();
    mql = installed.mql;
    restoreMatchMedia = installed.restore;
  });

  afterEach(() => {
    restoreMatchMedia();
    vi.restoreAllMocks();
  });

  test("queries matchMedia with the mobile breakpoint", () => {
    setInnerWidth(1024);
    renderHook(() => useIsMobile());
    expect(mql.media).toBe(`(max-width: ${MOBILE_BREAKPOINT - 1}px)`);
    expect(mql.addEventListener).toHaveBeenCalledWith(
      "change",
      expect.any(Function),
    );
  });

  test("returns false for desktop widths", () => {
    setInnerWidth(1024);
    const { result } = renderHook(() => useIsMobile());
    expect(result.current).toBe(false);
  });

  test("returns true for widths under the breakpoint", () => {
    setInnerWidth(500);
    const { result } = renderHook(() => useIsMobile());
    expect(result.current).toBe(true);
  });

  test("returns true exactly at one pixel below the breakpoint", () => {
    setInnerWidth(MOBILE_BREAKPOINT - 1);
    const { result } = renderHook(() => useIsMobile());
    expect(result.current).toBe(true);
  });

  test("returns false at the breakpoint itself", () => {
    setInnerWidth(MOBILE_BREAKPOINT);
    const { result } = renderHook(() => useIsMobile());
    expect(result.current).toBe(false);
  });

  test("updates when the media query reports a change", () => {
    setInnerWidth(1024);
    const { result } = renderHook(() => useIsMobile());
    expect(result.current).toBe(false);

    act(() => {
      setInnerWidth(500);
      mql.__fire(true);
    });
    expect(result.current).toBe(true);

    act(() => {
      setInnerWidth(1200);
      mql.__fire(false);
    });
    expect(result.current).toBe(false);
  });

  test("removes the change listener on unmount", () => {
    setInnerWidth(1024);
    const { unmount } = renderHook(() => useIsMobile());
    expect(mql.addEventListener).toHaveBeenCalledOnce();
    unmount();
    expect(mql.removeEventListener).toHaveBeenCalledWith(
      "change",
      expect.any(Function),
    );
  });
});
