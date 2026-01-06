import { createFileRoute, Link } from "@tanstack/react-router";

export const Route = createFileRoute("/discounts")({ component: Discounts });

function Discounts() {
  return (
    <div className="container mx-auto px-4 py-12">
      <h1 className="mb-6">Discounts</h1>
      <p className="mb-6">These companies offer discounts to our members.</p>
      <ul className="list-disc space-y-2 pl-6">
        <li>
          <Link
            to="/discounts/digitalocean"
            className="text-blue-600 hover:underline"
          >
            DigitalOcean
          </Link>{" "}
          (free hosting credits)
        </li>
      </ul>
    </div>
  );
}
