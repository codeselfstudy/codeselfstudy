import { Link } from "@tanstack/react-router";
import { SiGithub } from "@icons-pack/react-simple-icons";

import { CurrentYear } from "./current-year";

function FooterLink({
  to,
  children,
}: {
  to: string;
  children: React.ReactNode;
}) {
  return (
    <li>
      <Link className="text-gray-300" to={to}>
        {children}
      </Link>
    </li>
  );
}

export function Footer() {
  return (
    <footer className="bg-gray-900 py-6 pb-16 text-gray-400">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 gap-8 md:grid-cols-3">
          <div>
            <ul className="list-none space-y-1 pl-0">
              <FooterLink to="/contact">Contact Us</FooterLink>
              <FooterLink to="/learn">Learn How to Code</FooterLink>
              <FooterLink to="/puzzles">Coding Puzzles</FooterLink>
            </ul>
          </div>
          <div>
            <ul className="list-none space-y-1 pl-0">
              <FooterLink to="/jobs">Jobs</FooterLink>
              <FooterLink to="/about">About</FooterLink>
              <FooterLink to="/events">Events</FooterLink>
            </ul>
          </div>
          <div>
            <div className="flex gap-4">
              <a
                href="https://github.com/codeselfstudy/codeselfstudy"
                className="text-gray-400 transition-colors hover:text-white"
                aria-label="GitHub"
                target="_blank"
                rel="noopener noreferrer"
              >
                <SiGithub size={32} />
              </a>
            </div>
          </div>
        </div>
        <div className="mt-8 border-t border-gray-700 pt-8 text-center text-sm md:text-left">
          <p>
            &copy; <CurrentYear />{" "}
            <Link to="/" className="text-gray-300 hover:text-white">
              Code Self Study
            </Link>{" "}
            &bull; Programming community in the San Francisco Bay Area,
            California
          </p>
          <p className="mt-2"></p>
        </div>
      </div>
    </footer>
  );
}
