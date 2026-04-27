import type { ReactNode } from "react";

type LabelColor = "green" | "cyan" | "magenta" | "yellow";
type Variant = "default" | "full" | "last";

const LABEL_COLOR_VAR: Record<LabelColor, string> = {
  green: "var(--term-accent)",
  cyan: "var(--term-cyan)",
  magenta: "var(--term-magenta)",
  yellow: "var(--term-yellow)",
};

type TermPaneProps = {
  index: number;
  label: string;
  labelColor?: LabelColor;
  /** Right-aligned hint text in the pane bar (e.g. "↑↓ scroll · / filter"). */
  keys?: string;
  /** "full" spans both grid columns; "last" removes the right border. */
  variant?: Variant;
  children: ReactNode;
};

export function TermPane({
  index,
  label,
  labelColor = "green",
  keys,
  variant = "default",
  children,
}: TermPaneProps) {
  const noRightBorder = variant === "last" || variant === "full";
  return (
    <section
      style={{
        background: "var(--term-pane)",
        borderRightStyle: "solid",
        borderRightWidth: noRightBorder ? 0 : 1,
        borderRightColor: "var(--term-border)",
        borderBottom: "1px solid var(--term-border)",
        gridColumn: variant === "full" ? "1 / -1" : undefined,
      }}
    >
      <header
        style={{
          display: "flex",
          alignItems: "center",
          gap: 6,
          padding: "4px 10px",
          color: "var(--term-fg-dim)",
          fontSize: 12,
          borderBottom: "1px dashed var(--term-border)",
          background: "var(--term-pane-2)",
        }}
      >
        <span>{index}</span>
        <span style={{ color: LABEL_COLOR_VAR[labelColor] }}>{label}</span>
        <span
          aria-hidden="true"
          style={{
            flex: 1,
            borderBottom: "1px dashed var(--term-fg-mute)",
            margin: "0 4px",
            height: 0,
            transform: "translateY(2px)",
          }}
        />
        {keys ? (
          <span style={{ color: "var(--term-fg-mute)" }}>{keys}</span>
        ) : null}
      </header>
      <div style={{ padding: "12px 14px" }}>{children}</div>
    </section>
  );
}
