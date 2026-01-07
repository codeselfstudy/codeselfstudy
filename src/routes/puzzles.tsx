import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { PuzzlesContent } from "@/content/copyrighted/puzzles-content";
import { createMetadata } from "@/lib/metadata";

export const Route = createFileRoute("/puzzles")({
  component: Puzzles,

  head: () => ({
    meta: createMetadata({
      title: "Programming Puzzles",
      description: "Programming puzzles for the Code Self Study group",
    }),
  }),
});

function Puzzles() {
  return (
    <PageWrapper>
      <PuzzlesContent />
    </PageWrapper>
  );
}
