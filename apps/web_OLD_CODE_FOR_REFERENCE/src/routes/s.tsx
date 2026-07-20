import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { createMetadata } from "@/lib/metadata";
import { SContent } from "@/content/copyrighted/s-content";

export const Route = createFileRoute("/s")({
  component: RouteComponent,
  head: () => {
    return {
      meta: createMetadata({
        title: "Join our Slack",
        description: "Join our Slack to get started",
        isNoIndex: true,
      }),
    };
  },
});

function RouteComponent() {
  return (
    <PageWrapper>
      <SContent />
    </PageWrapper>
  );
}
