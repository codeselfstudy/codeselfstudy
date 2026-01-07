import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { ForumContent } from "@/content/copyrighted/forum-content";

export const Route = createFileRoute("/forum")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <PageWrapper>
      <ForumContent />
    </PageWrapper>
  );
}
