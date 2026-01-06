import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/PageWrapper";

export const Route = createFileRoute("/puzzles")({
  component: Puzzles,
});

function Puzzles() {
  return (
    <PageWrapper>
      <h1>Puzzles</h1>
      <p>A page of coding puzzles is coming soon!</p>
    </PageWrapper>
  );
}
