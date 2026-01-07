import { cn } from "@/lib/utils";
import { buttonVariants } from "@/components/ui/button";

export function CodewarsContent() {
  return (
    <>
      <h1>Codewars Clan</h1>
      <p>
        Click the button below to join our Codewars clan. If you aren't logged
        into Codewars, it will ask you to log in.
      </p>
      <p>
        <a
          href="https://www.codewars.com/r/0JGb7w"
          className={cn(buttonVariants({ size: "lg" }))}
          target="_blank"
          rel="noopener noreferrer"
        >
          Join our Codewars Clan
        </a>
      </p>
    </>
  );
}
