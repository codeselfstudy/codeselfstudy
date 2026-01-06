import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/puzzles")({
  component: Puzzles,
});

function Puzzles() {
  return (
    <div className="container mx-auto px-4 py-12">
      <h1 className="mb-6 text-4xl font-bold text-[#363636]">Puzzles</h1>
      <p className="text-lg text-[#363636]">
        A page of coding puzzles is coming soon!
      </p>
    </div>
  );
}
