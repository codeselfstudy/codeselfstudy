import { createFileRoute } from "@tanstack/react-router";
import { HomeContent } from "@/content/copyrighted/home-content";
import { createMetadata } from "@/lib/metadata";

export const Route = createFileRoute("/")({
  component: App,

  head: () => ({
    meta: createMetadata({
      title: "Code Self Study",
      description:
        "Welcome to Code Self Study: programming meetup group in the San Francisco Bay Area",
    }),
  }),
});

function App() {
  return <HomeContent />;
}
