import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/PageWrapper";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/discounts/digitalocean")({
  component: DigitalOcean,
});

function DigitalOcean() {
  return (
    <PageWrapper>
      <h1 className="mb-6">DigitalOcean</h1>
      <p className="mb-6">
        Digital Ocean hosts parts of our website and offers $50 credit on their
        hosting plans. Sign up with the link below to claim your $50 credit.
      </p>
      <blockquote className="mb-6 border-l-4 border-gray-300 pl-4 italic">
        "Simple Cloud Hosting, Built for Developers."
      </blockquote>
      <a
        href="https://m.do.co/c/974e5be9397c"
        className={cn(buttonVariants({ size: "lg" }))}
      >
        Sign up for $50 hosting credit
      </a>
    </PageWrapper>
  );
}
