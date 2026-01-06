import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/PageWrapper";
import { CodewarsContent } from "@/content/copyrighted/codewars-content";

export const Route = createFileRoute("/codewars")({ component: Codewars });

function Codewars() {
  return (
    <PageWrapper>
      <CodewarsContent />
    </PageWrapper>
  );
}
