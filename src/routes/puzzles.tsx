import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/PageWrapper";
import { PuzzlesContent } from "@/content/copyrighted/puzzles-content";

export const Route = createFileRoute("/puzzles")({
  component: Puzzles,
});

function Puzzles() {
  return (
    <PageWrapper>
      <PuzzlesContent />
    </PageWrapper>
  );
}
