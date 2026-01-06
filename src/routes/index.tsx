import { createFileRoute } from "@tanstack/react-router";
import { HomeContent } from "@/content/copyrighted/home-content";

export const Route = createFileRoute("/")({ component: App });

function App() {
  return <HomeContent />;
}
