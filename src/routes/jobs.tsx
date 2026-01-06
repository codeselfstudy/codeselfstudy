import { createFileRoute, Link } from "@tanstack/react-router";

export const Route = createFileRoute("/jobs")({ component: Jobs });

function Jobs() {
  return (
    <div className="container mx-auto px-4 py-12">
      <h1>Programming Jobs</h1>
      <p className="mb-6 max-w-2xl">
        This page is for tech jobs in the San Francisco Bay Area and elsewhere
        (if the employees are able to work remotely from the SF Bay Area).
      </p>
      <p>
        It's free to list a job here &mdash; just{" "}
        <Link to="/contact">send us a message</Link>.
      </p>
    </div>
  );
}
