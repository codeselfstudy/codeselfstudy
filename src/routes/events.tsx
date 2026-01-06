import { PageWrapper } from "@/components/PageWrapper";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/events")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <PageWrapper>
      <h1 className="mb-6">Events</h1>
      <p>
        Check our{" "}
        <a href="https://www.meetup.com/codeselfstudy/">Meetup group</a> for
        upcoming events.
      </p>
    </PageWrapper>
  );
}
