import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { CreditsContent } from "@/content/copyrighted/credits-content";

export const Route = createFileRoute("/credits")({ component: Credits });

function Credits() {
  return (
    <PageWrapper>
      <CreditsContent />
    </PageWrapper>
  );
}
