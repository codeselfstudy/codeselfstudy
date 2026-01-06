import { PageWrapper } from "@/components/PageWrapper";
import { createFileRoute } from "@tanstack/react-router";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/about")({ component: About });

function About() {
  return (
    <PageWrapper>
      <h1>About Us</h1>
      <p className="mb-6 max-w-2xl">
        We organize several tech meetup groups in Berkeley, California with over
        5,000 members in total. We also organize events around the San Francisco
        Bay Area.
      </p>
      <a
        href="https://www.meetup.com/codeselfstudy/"
        target="_blank"
        rel="noopener noreferrer nofollow"
        className={cn(buttonVariants({ size: "lg" }))}
      >
        Join the Meetup Group
      </a>
    </PageWrapper>
  );
}
