import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { CreditsContent } from "@/content/copyrighted/credits-content";
import { createMetadata } from "@/lib/metadata";

export const Route = createFileRoute("/credits")({
  component: Credits,

  head: () => ({
    meta: createMetadata({
      title: "Website Credits",
      description: "Website credits",
    }),
  }),
});

function Credits() {
  return (
    <PageWrapper>
      <CreditsContent />
    </PageWrapper>
  );
}
