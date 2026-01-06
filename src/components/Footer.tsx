import { Link } from "@tanstack/react-router";
import CurrentYear from "./CurrentYear";

export default function Footer() {
  return (
    <footer className="bg-[#333] py-[54px] pb-[108px] font-['Open_Sans'] text-[#bbb]">
      <div className="mx-auto max-w-[1344px] px-4">
        <div className="grid grid-cols-1 gap-8 md:grid-cols-3">
          <div>
            <h3 className="mb-4 text-lg font-bold text-white">About</h3>
            <p className="mb-4">
              Currently this is a rebuild of the original site.
            </p>
          </div>
          <div>
            <h3 className="mb-4 text-lg font-bold text-white">Links</h3>
            <ul className="space-y-2">
              <li>
                <Link to="/contact" className="hover:text-white">
                  Contact
                </Link>
              </li>
              <li>
                <Link to="/jobs" className="hover:text-white">
                  Jobs
                </Link>
              </li>
              <li>
                <Link to="/blog" className="hover:text-white">
                  Blog
                </Link>
              </li>
            </ul>
          </div>
          <div>
            <h3 className="mb-4 text-lg font-bold text-white">Social</h3>
            <ul className="space-y-2">
              <li>
                <a
                  href="https://github.com/codeselfstudy"
                  className="hover:text-white"
                >
                  GitHub
                </a>
              </li>
            </ul>
          </div>
        </div>
        <div className="mt-8 border-t border-[#444] pt-8 text-center md:text-left">
          &copy; <CurrentYear /> Code Self Study
        </div>
      </div>
    </footer>
  );
}
