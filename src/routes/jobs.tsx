import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { JobsContent } from "@/content/copyrighted/jobs-content";
import { createMetadata } from "@/lib/metadata";

export const Route = createFileRoute("/jobs")({
  component: Jobs,

  head: () => ({
    meta: createMetadata({
      title: "Job Opportunities",
      description:
        "Find job opportunities in the San Francisco Bay Area and elsewhere",
    }),
  }),
});

function Jobs() {
  return (
    <PageWrapper>
      <JobsContent />
    </PageWrapper>
  );
}
