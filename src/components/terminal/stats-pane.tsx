import { TermPane } from "./term-pane";
import { Meter } from "./meter";
import type { TermColor } from "@/content/mock-home-data";
import { STATS } from "@/content/mock-home-data";

const NET_COLOR_VAR: Record<TermColor, string> = {
  green: "var(--term-accent)",
  cyan: "var(--term-cyan)",
  yellow: "var(--term-yellow)",
  magenta: "var(--term-magenta)",
};

const COMMUNITY_VALUE_COLOR: Record<"green" | "cyan" | "magenta", string> = {
  green: "var(--term-accent)",
  cyan: "var(--term-cyan)",
  magenta: "var(--term-magenta)",
};

type StatsPaneProps = {
  stats?: typeof STATS;
};

export function StatsPane({ stats = STATS }: StatsPaneProps = {}) {
  return (
    <TermPane
      index={1}
      label="~/stats"
      labelColor="cyan"
      keys="e expand"
      variant="last"
    >
      <div style={{ color: "var(--term-fg-dim)", marginBottom: 6 }}>
        ── community ──────────────────────
      </div>
      {stats.community.map((row) => (
        <div
          key={row.label}
          style={{
            display: "flex",
            gap: 10,
            alignItems: "center",
            padding: "2px 0",
          }}
        >
          <span style={{ width: 70, color: "var(--term-fg-dim)" }}>
            {row.label}
          </span>
          <Meter
            value={row.meterPercent}
            color={row.color}
            ariaLabel={row.label}
          />
          <span
            style={{
              marginLeft: "auto",
              color: COMMUNITY_VALUE_COLOR[row.color],
            }}
          >
            {row.value}
          </span>
        </div>
      ))}
      <div style={{ color: "var(--term-fg-dim)", margin: "12px 0 6px" }}>
        ── net · this week ────────────────
      </div>
      {stats.netThisWeek.map((row) => (
        <div key={row.label} style={{ color: "var(--term-fg)" }}>
          <span aria-hidden="true">{row.icon} </span>
          <span style={{ color: NET_COLOR_VAR[row.color] }}>
            {row.value}
          </span>{" "}
          {row.label}
        </div>
      ))}
      <div style={{ color: "var(--term-fg-dim)", margin: "12px 0 4px" }}>
        ── load avg ───────────────────────
      </div>
      <pre
        style={{
          whiteSpace: "pre",
          color: "var(--term-fg-dim)",
          margin: 0,
          fontSize: 11,
          lineHeight: 1.05,
          fontFamily: "inherit",
        }}
      >
        <span style={{ color: "var(--term-accent)" }}>
          {stats.loadAvgAscii}
        </span>
        {"\n"}
        {stats.loadAvgCaption}
      </pre>
    </TermPane>
  );
}
