import { TermPane } from "./term-pane";
import { BlinkingCursor } from "./blinking-cursor";
import { WELCOME_COPY } from "@/content/mock-home-data";

type WelcomePaneProps = {
  copy?: typeof WELCOME_COPY;
};

export function WelcomePane({ copy = WELCOME_COPY }: WelcomePaneProps = {}) {
  const [leadBefore, leadAfter] = copy.lead.split(copy.leadHighlight);

  return (
    <TermPane
      index={2}
      label="~/welcome"
      labelColor="yellow"
      keys="prefix + ? help"
      variant="full"
    >
      <pre
        style={{
          whiteSpace: "pre",
          margin: 0,
          fontSize: 11,
          lineHeight: 1.05,
          color: "var(--term-fg-dim)",
          fontFamily: "inherit",
        }}
      >
        {"  "}
        <span style={{ color: "var(--term-accent)" }}>{copy.title}</span>
        {" · "}
        {copy.tagline}
      </pre>
      <p style={{ margin: "8px 0 4px", color: "var(--term-fg)" }}>
        {leadBefore}
        <b style={{ color: "var(--term-yellow)" }}>{copy.leadHighlight}</b>
        {leadAfter}
        <BlinkingCursor />
      </p>
      <div style={{ color: "var(--term-fg-dim)" }}>what we do ::</div>
      <ul
        style={{
          display: "flex",
          flexWrap: "wrap",
          gap: "4px 10px",
          marginTop: 6,
          padding: 0,
          listStyle: "none",
        }}
      >
        {copy.whatWeDo.map((item) => (
          <li key={item} style={{ color: "var(--term-fg)" }}>
            <span aria-hidden="true" style={{ color: "var(--term-fg-mute)" }}>
              ▸{" "}
            </span>
            {item}
          </li>
        ))}
      </ul>
      <div
        style={{
          display: "flex",
          gap: 10,
          flexWrap: "wrap",
          marginTop: 14,
        }}
      >
        <a
          href={copy.cta.href}
          style={{
            display: "inline-block",
            padding: "4px 12px",
            border: "1px solid var(--term-accent)",
            color: "var(--term-accent)",
            textDecoration: "none",
            fontFamily: "inherit",
          }}
        >
          {copy.cta.label}
        </a>
      </div>
    </TermPane>
  );
}
