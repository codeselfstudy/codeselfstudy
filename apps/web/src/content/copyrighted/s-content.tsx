import { cn } from "@/lib/utils";
import { buttonVariants } from "@/components/ui/button";

export function SContent() {
  return (
    <>
      <h1>Join our Slack</h1>
      <p>
        <a
          target="_blank"
          rel="noopener noreferrer nofollow"
          className={cn(buttonVariants({ size: "lg" }))}
          href="https://join.slack.com/t/codeselfstudy/shared_invite/enQtNTgzNDg3MjYyNTgyLTZiYzQ0OTU0NDA1OTA1NjljZGRkYmYxZDVhNGE4YzM0NTVmMTBjYjEyNmFiZmFmYjBiNzczOGIyYzNjZjkxMDY"
        >
          Join our Slack
        </a>
      </p>
    </>
  );
}
