import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { DiscountsContent } from "@/content/copyrighted/discounts-content";

export const Route = createFileRoute("/discounts")({ component: Discounts });

function Discounts() {
  return (
    <PageWrapper>
      <DiscountsContent />
    </PageWrapper>
  );
}
