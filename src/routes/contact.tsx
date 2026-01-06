import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/PageWrapper";

export const Route = createFileRoute("/contact")({ component: Contact });

function Contact() {
  return (
    <PageWrapper>
      <h1>Contact Us</h1>
      <p className="mb-6">
        Send us an email:{" "}
        <a href="mailto:contact@codeselfstudy.com">contact@codeselfstudy.com</a>
        .
      </p>
    </PageWrapper>
  );
}
