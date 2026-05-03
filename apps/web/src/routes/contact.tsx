import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { ContactContent } from "@/content/copyrighted/contact-content";
import { createMetadata } from "@/lib/metadata";

export const Route = createFileRoute("/contact")({
  component: Contact,

  head: () => ({
    meta: createMetadata({
      title: "Contact",
      description: "Contact Code Self Study group",
    }),
  }),
});

function Contact() {
  return (
    <PageWrapper>
      <ContactContent />
    </PageWrapper>
  );
}
