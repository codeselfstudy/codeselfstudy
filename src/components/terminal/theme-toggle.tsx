import { useEffect, useState } from "react";

type Theme = "dark" | "light";

const STORAGE_KEY = "tmux-theme";

function applyTheme(theme: Theme) {
  if (typeof document === "undefined") return;
  if (theme === "light") {
    document.documentElement.setAttribute("data-theme", "light");
  } else {
    document.documentElement.removeAttribute("data-theme");
  }
}

function readStoredTheme(): Theme {
  if (typeof window === "undefined") return "dark";
  const v = window.localStorage.getItem(STORAGE_KEY);
  return v === "light" ? "light" : "dark";
}

export function ThemeToggle() {
  const [theme, setTheme] = useState<Theme>("dark");
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    const initial = readStoredTheme();
    setTheme(initial);
    applyTheme(initial);
    setMounted(true);
  }, []);

  function flip() {
    const next: Theme = theme === "light" ? "dark" : "light";
    setTheme(next);
    applyTheme(next);
    try {
      window.localStorage.setItem(STORAGE_KEY, next);
    } catch {
      /* localStorage may be unavailable (e.g. private mode); ignore. */
    }
  }

  return (
    <button
      type="button"
      onClick={flip}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          flip();
        }
      }}
      aria-label="Toggle light and dark mode"
      className="terminal-theme-toggle"
      style={{
        cursor: "pointer",
        userSelect: "none",
        color: "var(--term-fg-dim)",
        fontSize: 12,
        padding: "2px 8px",
        border: "1px solid var(--term-border)",
        borderRadius: 4,
        background: "transparent",
        fontFamily: "inherit",
      }}
    >
      ◐ <span>{mounted ? theme : "dark"}</span>
    </button>
  );
}
