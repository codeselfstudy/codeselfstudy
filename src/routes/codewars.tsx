import { PageWrapper } from "@/components/PageWrapper";
import { createFileRoute } from "@tanstack/react-router";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/codewars")({ component: Codewars });

function Codewars() {
  return (
    <PageWrapper>
      <h1 className="mb-6">Codewars Clan</h1>
      <p className="mb-6">
        Click the button below to join our Codewars clan. If you aren't logged
        into Codewars, it will ask you to log in.
      </p>
      <a
        href="https://www.codewars.com/r/0JGb7w"
        className={cn(buttonVariants({ size: "lg" }))}
      >
        Join our Codewars Clan
      </a>
    </PageWrapper>
  );
}
