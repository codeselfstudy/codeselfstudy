import { Link } from "@tanstack/react-router";

export function CreditsContent() {
  return (
    <>
      <h1>Credits</h1>
      <p>
        Please support us by using our{" "}
        <Link to="/discounts">our discount links</Link>.
      </p>
      <p>
        SVG backgrounds are from{" "}
        <a
          target="_blank"
          rel="noopener noreferrer"
          href="https://www.heropatterns.com/"
        >
          Hero Patterns
        </a>{" "}
        and used under the{" "}
        <a
          target="_blank"
          rel="noopener noreferrer"
          href="https://creativecommons.org/licenses/by/4.0/"
        >
          CC by 4.0 license
        </a>
        .
      </p>
      <p>
        The website framework is{" "}
        <a
          target="_blank"
          rel="noopener noreferrer"
          href="https://tanstack.com/"
        >
          TanStack Start
        </a>
        .
      </p>
    </>
  );
}
