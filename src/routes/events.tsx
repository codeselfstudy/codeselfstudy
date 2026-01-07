import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { EventsContent } from "@/content/copyrighted/events-content";
import { createMetadata } from "@/lib/metadata";

export const Route = createFileRoute("/events")({
  component: RouteComponent,

  head: () => ({
    meta: createMetadata({
      title: "Events for Programmers",
      description: "Code Self Study organizes events for Programmers",
    }),
  }),
});

function RouteComponent() {
  return (
    <PageWrapper>
      <EventsContent />
    </PageWrapper>
  );
}
