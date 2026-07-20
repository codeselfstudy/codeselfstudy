import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { ForumContent } from "@/content/copyrighted/forum-content";
import { createMetadata } from "@/lib/metadata";

export const Route = createFileRoute("/forum")({
  component: RouteComponent,
  head: () => {
    return {
      meta: createMetadata({
        title: "Programming Forum",
        description: "Join our free programming forum",
      }),
    };
  },
});

function RouteComponent() {
  return (
    <PageWrapper>
      <ForumContent />
    </PageWrapper>
  );
}
