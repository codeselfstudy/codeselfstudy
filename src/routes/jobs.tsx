import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { JobsContent } from "@/content/copyrighted/jobs-content";

export const Route = createFileRoute("/jobs")({ component: Jobs });

function Jobs() {
  return (
    <PageWrapper>
      <JobsContent />
    </PageWrapper>
  );
}
