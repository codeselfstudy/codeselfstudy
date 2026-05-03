import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { LearnContent } from "@/content/copyrighted/learn-content";
import { createMetadata } from "@/lib/metadata";

export const Route = createFileRoute("/learn")({
  component: RouteComponent,
  head: () => {
    return {
      meta: createMetadata({
        title: "Learn How to Code",
        description: "Learn how to code with the Code Self Study community",
      }),
    };
  },
});

function RouteComponent() {
  return (
    <PageWrapper>
      <LearnContent />
    </PageWrapper>
  );
}
