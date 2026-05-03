import { Link, useRouter } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { Button, buttonVariants } from "@/components/ui/button";

interface ErrorPageProps {
  error?: Error;
}

export function ErrorPage({ error }: ErrorPageProps) {
  const router = useRouter();

  return (
    <PageWrapper>
      <div className="mx-auto max-w-2xl text-center">
        <p className="text-muted-foreground text-sm font-semibold tracking-widest uppercase">
          500
        </p>
        <h1 className="mt-2 text-4xl font-bold tracking-tight sm:text-5xl">
          Something went wrong
        </h1>
        <p className="text-muted-foreground mt-4 text-base">
          An unexpected error occurred while loading this page. You can try
          again, or head back home.
        </p>
        {import.meta.env.DEV && error?.message && (
          <pre className="bg-muted text-muted-foreground mt-6 overflow-x-auto rounded-md p-4 text-left text-xs">
            {error.message}
          </pre>
        )}
        <div className="mt-8 flex items-center justify-center gap-3">
          <Button onClick={() => router.invalidate()}>Try again</Button>
          <Link to="/" className={buttonVariants({ variant: "outline" })}>
            Go home
          </Link>
        </div>
      </div>
    </PageWrapper>
  );
}
