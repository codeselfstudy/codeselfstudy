import "@testing-library/jest-dom/vitest";

// Base UI's Dialog (used by the Sheet in the navbar) touches a few browser
// APIs that jsdom does not implement. Provide minimal stubs so component
// tests can render and open the drawer.

if (!window.matchMedia) {
  window.matchMedia = ((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addEventListener() {},
    removeEventListener() {},
    addListener() {},
    removeListener() {},
    dispatchEvent() {
      return false;
    },
  })) as typeof window.matchMedia;
}

if (!("ResizeObserver" in globalThis)) {
  globalThis.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  } as unknown as typeof ResizeObserver;
}

for (const method of [
  "hasPointerCapture",
  "setPointerCapture",
  "releasePointerCapture",
  "scrollIntoView",
] as const) {
  if (!(method in Element.prototype)) {
    (Element.prototype as unknown as Record<string, () => void>)[method] =
      () => {};
  }
}
