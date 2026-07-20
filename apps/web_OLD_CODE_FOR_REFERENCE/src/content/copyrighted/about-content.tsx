import { cn } from "@/lib/utils";
import { buttonVariants } from "@/components/ui/button";

export function AboutContent() {
  return (
    <>
      <h1>About Us</h1>
      <p>
        We organize several tech meetup groups in Berkeley, California with over
        5,000 members in total. We also organize events around the San Francisco
        Bay Area.
      </p>
      <p>
        <a
          href="https://www.meetup.com/codeselfstudy/"
          target="_blank"
          rel="noopener noreferrer nofollow"
          className={cn(buttonVariants({ size: "lg" }))}
        >
          Join the Meetup Group
        </a>
      </p>
    </>
  );
}
