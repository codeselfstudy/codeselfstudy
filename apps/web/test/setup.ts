import "@testing-library/jest-dom/vitest";

// Node 26 puts `Storage`, `localStorage` and `sessionStorage` on the global
// object, and `localStorage` is `undefined` unless the process was started with
// --localstorage-file. Vitest's jsdom environment only copies a jsdom global
// across when the key is missing from the host global, so on Node 26 jsdom's
// real storage never lands and every `localStorage.clear()` in a hook throws --
// before Testing Library's auto-cleanup, which then leaks the DOM between
// tests. Put a minimal in-memory Storage back so the suite behaves the same on
// any Node version and under Bun. `Storage` is replaced too: the tests spy on
// `Storage.prototype`, so the class and the instances have to match.
//
// Reading the global has to be guarded as well as shape-checked: Node's Web
// Storage is experimental and today's `undefined` could become a throw, and a
// jsdom window on an opaque origin (`about:blank`, `file:`) throws SecurityError
// from the getter. Either way there is no usable storage, and letting the probe
// itself throw would take down every test file instead of just this hook.
function hostStorageIsUsable() {
  try {
    return typeof globalThis.localStorage?.clear === "function";
  } catch {
    return false;
  }
}

if (!hostStorageIsUsable()) {
  class MemoryStorage {
    #entries = new Map<string, string>();

    get length() {
      return this.#entries.size;
    }

    key(index: number) {
      return [...this.#entries.keys()][index] ?? null;
    }

    getItem(key: string) {
      return this.#entries.get(String(key)) ?? null;
    }

    setItem(key: string, value: string) {
      this.#entries.set(String(key), String(value));
    }

    removeItem(key: string) {
      this.#entries.delete(String(key));
    }

    clear() {
      this.#entries.clear();
    }
  }

  // Node's `localStorage` is an accessor, so a plain assignment would not stick.
  const define = (name: string, value: unknown) =>
    Object.defineProperty(globalThis, name, {
      configurable: true,
      writable: true,
      value,
    });

  define("Storage", MemoryStorage);
  define("localStorage", new MemoryStorage());
  define("sessionStorage", new MemoryStorage());
}

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
