import { Link } from "@tanstack/react-router";
import type { ReactNode } from "react";

export type NavLink = {
  to: string;
  label: string;
};

export const NAV_LINKS: Array<NavLink> = [
  { to: "/", label: "home" },
  { to: "/events", label: "events" },
  { to: "/learn", label: "learn" },
  { to: "/forum", label: "forum" },
  { to: "/jobs", label: "jobs" },
  { to: "/puzzles", label: "puzzles" },
  { to: "/tools", label: "tools" },
  { to: "/about", label: "about" },
];

type TermNavProps = {
  /** Slot for the theme toggle (kept as a slot so TermNav stays presentational). */
  themeToggle?: ReactNode;
  /** Override the link list (default: NAV_LINKS). */
  links?: Array<NavLink>;
};

export function TermNav({ themeToggle, links = NAV_LINKS }: TermNavProps) {
  return (
    <nav
      aria-label="primary"
      style={{
        display: "flex",
        gap: 8,
        flexWrap: "wrap",
        alignItems: "center",
        padding: "10px 14px",
        background: "var(--term-pane)",
        borderBottom: "1px solid var(--term-border)",
        fontFamily: "var(--font-terminal-mono)",
      }}
    >
      <span
        style={{
          color: "var(--term-fg)",
          fontSize: 13,
          marginRight: 12,
          letterSpacing: "0.04em",
        }}
      >
        <b style={{ color: "var(--term-accent)", fontWeight: 600 }}>
          codeselfstudy
        </b>
      </span>
      {links.map((link) => (
        // @ts-expect-error TanStack Router's typed routes are codegen'd; allow string `to`.
        <Link
          key={link.to}
          to={link.to}
          style={{
            fontSize: 12,
            color: "var(--term-fg-dim)",
            padding: "2px 10px",
            border: "1px solid var(--term-border)",
            borderRadius: 4,
            textDecoration: "none",
            letterSpacing: "0.04em",
          }}
        >
          {link.label}
        </Link>
      ))}
      {themeToggle ? (
        <span style={{ marginLeft: "auto" }}>{themeToggle}</span>
      ) : null}
    </nav>
  );
}
