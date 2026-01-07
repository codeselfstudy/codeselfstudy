import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { CodewarsContent } from "@/content/copyrighted/codewars-content";
import { createMetadata } from "@/lib/metadata";

export const Route = createFileRoute("/codewars")({
  component: Codewars,

  head: () => ({
    meta: createMetadata({
      title: "Join our Codewars Group",
      description: "Join our Codewars Group",
    }),
  }),
});

function Codewars() {
  return (
    <PageWrapper>
      <CodewarsContent />
    </PageWrapper>
  );
}
