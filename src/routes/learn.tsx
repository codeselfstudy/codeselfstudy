import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { LearnContent } from "@/content/copyrighted/learn-content";

export const Route = createFileRoute("/learn")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <PageWrapper>
      <LearnContent />
    </PageWrapper>
  );
}
