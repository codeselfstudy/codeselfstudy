import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { AboutContent } from "@/content/copyrighted/about-content";

export const Route = createFileRoute("/about")({ component: About });

function About() {
  return (
    <PageWrapper>
      <AboutContent />
    </PageWrapper>
  );
}
