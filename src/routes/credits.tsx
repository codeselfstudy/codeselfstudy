import { Link, createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/PageWrapper";

export const Route = createFileRoute("/credits")({ component: Credits });

function Credits() {
  return (
    <PageWrapper>
      <h1 className="mb-6">Credits</h1>
      <p className="mb-6">
        Please support us by using our{" "}
        <Link to="/discounts">our discount links</Link>.
      </p>
      <p className="mb-6">
        SVG backgrounds are from{" "}
        <a href="https://www.heropatterns.com/">Hero Patterns</a> and used under
        the{" "}
        <a href="https://creativecommons.org/licenses/by/4.0/">
          CC by 4.0 license
        </a>
        .
      </p>
      <p className="mb-6">
        The website framework is{" "}
        <a href="https://tanstack.com/">TanStack Start</a>.
      </p>
    </PageWrapper>
  );
}
