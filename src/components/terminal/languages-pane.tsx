import { TermPane } from "./term-pane";
import { AsciiBar } from "./ascii-bar";
import type { Language, TermColor } from "@/content/mock-home-data";
import { LANGUAGES, LANGUAGES_OVERFLOW_LABEL } from "@/content/mock-home-data";

const PCT_COLOR_VAR: Record<TermColor, string> = {
  green: "var(--term-accent)",
  cyan: "var(--term-cyan)",
  yellow: "var(--term-yellow)",
  magenta: "var(--term-magenta)",
};

const BAR_WIDTH = 46;

type LanguagesPaneProps = {
  languages?: Array<Language>;
  overflowLabel?: string;
};

export function LanguagesPane({
  languages = LANGUAGES,
  overflowLabel = LANGUAGES_OVERFLOW_LABEL,
}: LanguagesPaneProps = {}) {
  return (
    <TermPane index={0} label="~/languages" keys="↑↓ scroll · / filter">
      <div style={{ color: "var(--term-fg-dim)", marginBottom: 6 }}>
        ── languages used by members ──────────────────────
      </div>
      {languages.map((lang) => (
        <div
          key={lang.name}
          style={{
            display: "grid",
            gridTemplateColumns: "90px 1fr 60px",
            gap: 10,
            alignItems: "center",
            padding: "1px 0",
          }}
        >
          <span style={{ color: "var(--term-fg)" }}>{lang.name}</span>
          <AsciiBar value={lang.percent} width={BAR_WIDTH} color={lang.color} />
          <span
            style={{
              textAlign: "right",
              color: PCT_COLOR_VAR[lang.color],
            }}
          >
            {lang.percent}%
          </span>
        </div>
      ))}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "90px 1fr 60px",
          gap: 10,
          alignItems: "center",
          padding: "1px 0",
        }}
      >
        <span style={{ color: "var(--term-fg)" }}>{overflowLabel}</span>
        <span
          aria-hidden="true"
          style={{
            color: "var(--term-fg-mute)",
            whiteSpace: "pre",
            letterSpacing: "-1px",
          }}
        >
          {"░".repeat(BAR_WIDTH)}
        </span>
        <span style={{ textAlign: "right", color: "var(--term-fg-dim)" }}>
          ··
        </span>
      </div>
    </TermPane>
  );
}
