import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, test } from "vitest";

const cssPath = fileURLToPath(new URL("./styles.css", import.meta.url));
const css = readFileSync(cssPath, "utf-8");

const REQUIRED_TOKENS = [
  "--term-bg",
  "--term-pane",
  "--term-pane-2",
  "--term-fg",
  "--term-fg-dim",
  "--term-fg-mute",
  "--term-border",
  "--term-border-hi",
  "--term-accent",
  "--term-magenta",
  "--term-cyan",
  "--term-yellow",
  "--term-red",
  "--term-status-bg",
  "--term-status-fg",
  "--term-status-active",
  "--term-dot-red",
  "--term-dot-yellow",
  "--term-dot-green",
  "--term-glow",
  "--term-scanline",
];

function blockFor(selector: string): string {
  const re = new RegExp(
    `${selector.replace(/[\\^$.*+?()[\]{}|]/g, "\\$&")}\\s*\\{([^}]*)\\}`,
    "m"
  );
  const m = css.match(re);
  if (!m) throw new Error(`selector not found in styles.css: ${selector}`);
  return m[1];
}

describe("terminal theme tokens", () => {
  test(":root defines every required terminal token", () => {
    const root = blockFor(":root");
    for (const token of REQUIRED_TOKENS) {
      expect(root, `missing ${token} in :root`).toContain(token);
    }
  });

  test(`[data-theme="light"] overrides every required terminal token`, () => {
    const light = blockFor(`[data-theme="light"]`);
    for (const token of REQUIRED_TOKENS) {
      expect(light, `missing ${token} in [data-theme="light"]`).toContain(
        token
      );
    }
  });

  test("dark and light values for --term-accent differ", () => {
    const root = blockFor(":root");
    const light = blockFor(`[data-theme="light"]`);
    const re = /--term-accent:\s*([^;]+);/;
    const dark = root.match(re)?.[1].trim();
    const lightVal = light.match(re)?.[1].trim();
    expect(dark).toBeTruthy();
    expect(lightVal).toBeTruthy();
    expect(dark).not.toBe(lightVal);
  });

  test("monospace font variable is registered", () => {
    expect(css).toMatch(/--font-terminal-mono:\s*"JetBrains Mono"/);
  });

  test("@keyframes blink is defined", () => {
    expect(css).toMatch(/@keyframes\s+blink\s*\{[^}]*opacity:\s*0/);
  });

  test(".terminal-shell utility applies scanline + mono font", () => {
    const re = /\.terminal-shell\s*\{([^}]*)\}/m;
    const block = css.match(re)?.[1];
    expect(block).toBeTruthy();
    expect(block).toContain("var(--font-terminal-mono)");
    expect(block).toContain("repeating-linear-gradient");
    expect(block).toContain("var(--term-scanline)");
    expect(block).toContain("text-shadow");
  });
});
