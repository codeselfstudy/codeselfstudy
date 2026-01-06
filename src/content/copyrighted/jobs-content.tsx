import { Link } from "@tanstack/react-router";

export function JobsContent() {
  return (
    <>
      <h1>Programming Jobs</h1>
      <p>
        This page is for tech jobs in the San Francisco Bay Area and elsewhere
        (if the employees are able to work remotely from the SF Bay Area).
      </p>
      <p>
        It's free to list a job here &mdash; just{" "}
        <Link to="/contact">send us a message</Link>.
      </p>
    </>
  );
}
