import { Link } from "@tanstack/react-router";
import { Github } from "lucide-react";
import CurrentYear from "./CurrentYear";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export default function Footer() {
  return (
    <footer className="bg-[#333] py-[54px] pb-[108px] font-['Open_Sans'] text-[#bbb]">
      <div className="mx-auto max-w-[1344px] px-4">
        <div className="grid grid-cols-1 gap-8 md:grid-cols-3">
          <div>
            <ul className="space-y-1">
              <li>
                <Link
                  to="/contact"
                  className={cn(
                    buttonVariants({ variant: "link" }),
                    "h-auto p-0 text-[#bbb] hover:text-white"
                  )}
                >
                  Contact Us
                </Link>
              </li>
              <li>
                <Link
                  to="/learn"
                  className={cn(
                    buttonVariants({ variant: "link" }),
                    "h-auto p-0 text-[#bbb] hover:text-white"
                  )}
                >
                  Learn How to Code
                </Link>
              </li>
              <li>
                <Link
                  to="/puzzles"
                  className={cn(
                    buttonVariants({ variant: "link" }),
                    "h-auto p-0 text-[#bbb] hover:text-white"
                  )}
                >
                  Coding Puzzles
                </Link>
              </li>
              <li>
                <a
                  href="https://blog.codeselfstudy.com/"
                  className={cn(
                    buttonVariants({ variant: "link" }),
                    "h-auto p-0 text-[#bbb] hover:text-white"
                  )}
                >
                  Blog
                </a>
              </li>
              <li>
                <Link
                  to="/school"
                  className={cn(
                    buttonVariants({ variant: "link" }),
                    "h-auto p-0 text-[#bbb] hover:text-white"
                  )}
                >
                  School
                </Link>
              </li>
            </ul>
          </div>
          <div>
            <ul className="space-y-1">
              <li>
                <Link
                  to="/jobs"
                  className={cn(
                    buttonVariants({ variant: "link" }),
                    "h-auto p-0 text-[#bbb] hover:text-white"
                  )}
                >
                  Jobs
                </Link>
              </li>
              <li>
                <Link
                  to="/about"
                  className={cn(
                    buttonVariants({ variant: "link" }),
                    "h-auto p-0 text-[#bbb] hover:text-white"
                  )}
                >
                  About
                </Link>
              </li>
              <li>
                <Link
                  to="/events"
                  className={cn(
                    buttonVariants({ variant: "link" }),
                    "h-auto p-0 text-[#bbb] hover:text-white"
                  )}
                >
                  Events
                </Link>
              </li>
            </ul>
          </div>
          <div>
            <div className="flex gap-4">
              <a
                href="https://github.com/codeselfstudy/codeselfstudy"
                className="text-[#bbb] transition-colors hover:text-white"
                aria-label="GitHub"
              >
                <Github size={32} />
              </a>
            </div>
          </div>
        </div>
        <div className="mt-8 border-t border-[#444] pt-8 text-center text-sm md:text-left">
          <p>
            <Link to="/" className="hover:text-white">
              Code Self Study
            </Link>{" "}
            &bull; Programming community in Berkeley &amp; San Francisco Bay
            Area, California
          </p>
          <p className="mt-2">
            &copy; <CurrentYear /> Code Self Study
          </p>
        </div>
      </div>
    </footer>
  );
}
