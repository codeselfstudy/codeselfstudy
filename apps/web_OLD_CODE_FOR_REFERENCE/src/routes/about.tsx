import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { AboutContent } from "@/content/copyrighted/about-content";
import { createMetadata } from "@/lib/metadata";

export const Route = createFileRoute("/about")({
  component: About,
  head: () => ({
    meta: createMetadata({
      title: "About",
      description: "About the Code Self Study website",
    }),
  }),
});

function About() {
  return (
    <PageWrapper>
      <AboutContent />
    </PageWrapper>
  );
}
