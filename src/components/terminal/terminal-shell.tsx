import type { ReactNode } from "react";

type TerminalShellProps = {
  /** The navboxes/top nav row. Render <TermNav /> here from the call site. */
  nav: ReactNode;
  /** The bottom tmux status bar. Render <TmuxStatusBar ... /> here. */
  statusBar: ReactNode;
  children: ReactNode;
};

/**
 * Outer terminal-pane wrapper used by the root layout to apply the mockup-68
 * shell (scanlines, mono font, full-bleed pane). Pure structural component;
 * <TermNav>, <TmuxStatusBar>, and the page body are passed in as slots.
 */
export function TerminalShell({
  nav,
  statusBar,
  children,
}: TerminalShellProps) {
  return (
    <div
      className="terminal-shell"
      style={{ display: "flex", flexDirection: "column" }}
    >
      {nav}
      <main style={{ flex: 1, display: "flex", flexDirection: "column" }}>
        {children}
      </main>
      {statusBar}
    </div>
  );
}
