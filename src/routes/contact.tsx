import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/contact")({ component: Contact });

function Contact() {
  return (
    <div className="container mx-auto px-4 py-12">
      <h1>Contact Us</h1>
      <p className="mb-6">
        Send us an email:{" "}
        <a
          href="mailto:contact@codeselfstudy.com"
          className="text-blue-600 hover:underline"
        >
          contact@codeselfstudy.com
        </a>
        .
      </p>
    </div>
  );
}
