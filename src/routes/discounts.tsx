import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { DiscountsContent } from "@/content/copyrighted/discounts-content";
import { createMetadata } from "@/lib/metadata";

export const Route = createFileRoute("/discounts")({
  component: Discounts,

  head: () => ({
    meta: createMetadata({
      title: "Discounts for the Group",
      description: "Discounts for Code Self Study group members",
    }),
  }),
});

function Discounts() {
  return (
    <PageWrapper>
      <DiscountsContent />
    </PageWrapper>
  );
}
