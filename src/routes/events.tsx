import { createFileRoute } from "@tanstack/react-router";
import { PageWrapper } from "@/components/page-wrapper";
import { EventsContent } from "@/content/copyrighted/events-content";

export const Route = createFileRoute("/events")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <PageWrapper>
      <EventsContent />
    </PageWrapper>
  );
}
