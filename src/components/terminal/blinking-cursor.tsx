export function BlinkingCursor() {
  return (
    <span
      aria-hidden="true"
      style={{
        color: "var(--term-accent)",
        animation: "blink 1.1s steps(1) infinite",
      }}
    >
      ▮
    </span>
  );
}
