type Color = "green" | "cyan" | "yellow" | "magenta";

const COLOR_VAR: Record<Color, string> = {
  green: "var(--term-accent)",
  cyan: "var(--term-cyan)",
  yellow: "var(--term-yellow)",
  magenta: "var(--term-magenta)",
};

type AsciiBarProps = {
  /** 0–100, clamped. */
  value: number;
  /** Total character width of the bar. */
  width: number;
  color?: Color;
};

export function AsciiBar({ value, width, color = "green" }: AsciiBarProps) {
  const clamped = Math.max(0, Math.min(100, value));
  const filled = Math.round((clamped / 100) * width);
  const empty = width - filled;
  return (
    <span
      aria-hidden="true"
      style={{
        color: COLOR_VAR[color],
        whiteSpace: "pre",
        overflow: "hidden",
        letterSpacing: "-1px",
        fontFamily: "inherit",
      }}
    >
      {"█".repeat(filled)}
      {"░".repeat(empty)}
    </span>
  );
}
