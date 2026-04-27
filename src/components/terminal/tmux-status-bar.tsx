import { useEffect, useState } from "react";
import type { ReactNode } from "react";
import type { TmuxWindow } from "@/content/mock-home-data";

function pad(n: number): string {
  return String(n).padStart(2, "0");
}

function nowHHMM(): string {
  const d = new Date();
  return `${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

/** Live HH:MM clock; SSR-safe (renders empty placeholder until mounted). */
export function Clock() {
  const [text, setText] = useState<string>("");
  useEffect(() => {
    setText(nowHHMM());
    const id = window.setInterval(() => setText(nowHHMM()), 30_000);
    return () => window.clearInterval(id);
  }, []);
  return (
    <span aria-label="current time" suppressHydrationWarning>
      {text}
    </span>
  );
}

type TmuxStatusBarProps = {
  session?: string;
  windows: Array<TmuxWindow>;
  /** Index of the active window (0-based). */
  activeIndex: number;
  /** Right-aligned status content. Defaults to the right text + <Clock />. */
  right?: ReactNode;
};

export function TmuxStatusBar({
  session = "codeselfstudy",
  windows,
  activeIndex,
  right,
}: TmuxStatusBarProps) {
  return (
    <div
      role="status"
      aria-label="tmux status bar"
      style={{
        display: "flex",
        alignItems: "center",
        gap: 14,
        padding: "4px 12px",
        background: "var(--term-status-bg)",
        color: "var(--term-status-fg)",
        fontSize: 12,
        fontFamily: "var(--font-terminal-mono)",
      }}
    >
      <span
        style={{
          background: "var(--term-status-active)",
          color: "var(--term-status-bg)",
          padding: "0 6px",
          fontWeight: 600,
        }}
      >
        [{session}]
      </span>
      {windows.map((w, i) => {
        const active = i === activeIndex;
        return (
          <span
            key={w.id}
            aria-current={active ? "page" : undefined}
            style={{
              opacity: active ? 1 : 0.7,
              textDecoration: active ? "underline" : "none",
            }}
          >
            {w.id}:{w.label}
            {active ? "*" : ""}
          </span>
        );
      })}
      <span style={{ marginLeft: "auto", opacity: 0.85 }}>{right}</span>
    </div>
  );
}
