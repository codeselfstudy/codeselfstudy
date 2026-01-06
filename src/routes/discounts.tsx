import { createFileRoute, Link } from "@tanstack/react-router";

export const Route = createFileRoute("/discounts")({ component: Discounts });

function Discounts() {
  return (
    <div className="container mx-auto px-4 py-12">
      <h1 className="mb-6">Discounts</h1>
      <p className="mb-6">
        These companies offer discounts to our members. We may make small
        commissions or discounts when you use these links, which helps offset
        the expenses of running the group.
      </p>
      <ul className="list-disc space-y-2 pl-6">
        <li>
          <Link to="/discounts/digitalocean">DigitalOcean</Link> (free hosting
          credits)
        </li>
      </ul>
    </div>
  );
}
