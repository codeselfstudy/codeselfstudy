import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { ToolsContent } from "@/content/copyrighted/tools-content";

export const Route = createFileRoute("/tools")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <PageWrapper>
      <ToolsContent />
    </PageWrapper>
  );
}
