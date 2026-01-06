import { Link, createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/PageWrapper";

export const Route = createFileRoute("/jobs")({ component: Jobs });

function Jobs() {
  return (
    <PageWrapper>
      <h1>Programming Jobs</h1>
      <p className="mb-6 max-w-2xl">
        This page is for tech jobs in the San Francisco Bay Area and elsewhere
        (if the employees are able to work remotely from the SF Bay Area).
      </p>
      <p>
        It's free to list a job here &mdash; just{" "}
        <Link to="/contact">send us a message</Link>.
      </p>
    </PageWrapper>
  );
}
