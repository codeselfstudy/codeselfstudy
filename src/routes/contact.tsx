import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { ContactContent } from "@/content/copyrighted/contact-content";

export const Route = createFileRoute("/contact")({ component: Contact });

function Contact() {
  return (
    <PageWrapper>
      <ContactContent />
    </PageWrapper>
  );
}
