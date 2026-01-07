import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { ToolsContent } from "@/content/copyrighted/tools-content";
import { createMetadata } from "@/lib/metadata";

export const Route = createFileRoute("/tools")({
  component: RouteComponent,
  head: () => {
    return {
      meta: createMetadata({
        title: "Tools",
        description: "Tools built by the Code Self Study community",
      }),
    };
  },
});

function RouteComponent() {
  return (
    <PageWrapper>
      <ToolsContent />
    </PageWrapper>
  );
}
