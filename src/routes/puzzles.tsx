import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/puzzles")({
  component: Puzzles,
});

function Puzzles() {
  return (
    <div className="container mx-auto px-4 py-12">
      <h1>Puzzles</h1>
      <p>A page of coding puzzles is coming soon!</p>
    </div>
  );
}
