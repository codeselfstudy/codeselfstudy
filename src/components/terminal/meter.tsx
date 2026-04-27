type Color = "green" | "cyan" | "magenta";

const COLOR_VAR: Record<Color, string> = {
  green: "var(--term-accent)",
  cyan: "var(--term-cyan)",
  magenta: "var(--term-magenta)",
};

type MeterProps = {
  /** 0–100, clamped. */
  value: number;
  color?: Color;
  ariaLabel?: string;
};

export function Meter({ value, color = "green", ariaLabel }: MeterProps) {
  const clamped = Math.max(0, Math.min(100, value));
  const fill = COLOR_VAR[color];
  return (
    <span
      role="meter"
      aria-label={ariaLabel}
      aria-valuenow={clamped}
      aria-valuemin={0}
      aria-valuemax={100}
      style={{
        flex: 1,
        display: "block",
        height: 10,
        border: "1px solid var(--term-border)",
        background: "var(--term-pane-2)",
        position: "relative",
      }}
    >
      <i
        style={{
          display: "block",
          height: "100%",
          width: `${clamped}%`,
          background: fill,
          boxShadow: `0 0 6px ${fill}`,
        }}
      />
    </span>
  );
}
