import { Link } from "@tanstack/react-router";
import { Github } from "lucide-react";
import CurrentYear from "./CurrentYear";

export default function Footer() {
  return (
    <footer className="bg-gray-900 py-6 pb-16 text-gray-400">
      <div className="mx-auto max-w-7xl px-4">
        <div className="grid grid-cols-1 gap-8 md:grid-cols-3">
          <div>
            <ul className="space-y-1">
              <li>
                <Link to="/contact">Contact Us</Link>
              </li>
              <li>
                <Link to="/puzzles">Coding Puzzles</Link>
              </li>
            </ul>
          </div>
          <div>
            <ul className="space-y-1">
              <li>
                <Link to="/jobs">Jobs</Link>
              </li>
              <li>
                <Link to="/about">About</Link>
              </li>
              <li>
                <Link to="/events">Events</Link>
              </li>
            </ul>
          </div>
          <div>
            <div className="flex gap-4">
              <a
                href="https://github.com/codeselfstudy/codeselfstudy"
                className="text-gray-400 transition-colors hover:text-white"
                aria-label="GitHub"
              >
                <Github size={32} />
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
