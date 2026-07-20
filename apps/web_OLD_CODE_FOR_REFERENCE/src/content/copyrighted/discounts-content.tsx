import { cn } from "@/lib/utils";
import { buttonVariants } from "@/components/ui/button";

export function DiscountsContent() {
  return (
    <>
      <h1>Discounts</h1>
      <p>
        These companies offer discounts to our members. We may make small
        commissions or discounts when you use these links, which helps offset
        the expenses of running the group.
      </p>

      <h2>DigitalOcean</h2>

      <p>
        Digital Ocean hosts parts of our website and offers $50 credit on their
        hosting plans. Sign up with the link below to claim your $50 credit.
      </p>
      <blockquote>"Simple Cloud Hosting, Built for Developers."</blockquote>
      <a
        href="https://m.do.co/c/974e5be9397c"
        className={cn(buttonVariants({ size: "lg" }))}
        target="_blank"
        rel="noopener noreferrer nofollow"
      >
        Sign up for $200 hosting credit
      </a>
    </>
  );
}
