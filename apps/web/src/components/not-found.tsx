import { Link } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { buttonVariants } from "@/components/ui/button";

export function NotFound() {
  return (
    <PageWrapper>
      <div className="mx-auto max-w-2xl text-center">
        <p className="text-muted-foreground text-sm font-semibold tracking-widest uppercase">
          404
        </p>
        <h1 className="mt-2 text-4xl font-bold tracking-tight sm:text-5xl">
          Page not found
        </h1>
        <p className="text-muted-foreground mt-4 text-base">
          We couldn&apos;t find the page you were looking for. It may have been
          moved or no longer exists.
        </p>
        <div className="mt-8 flex items-center justify-center gap-3">
          <Link to="/" className={buttonVariants()}>
            Go home
          </Link>
          <Link
            to="/contact"
            className={buttonVariants({ variant: "outline" })}
          >
            Contact us
          </Link>
        </div>
      </div>
    </PageWrapper>
  );
}
