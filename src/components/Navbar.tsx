import { useState } from "react";

const LINKS = [
  { href: "#", label: "Home" },
  { href: "#", label: "About" },
  { href: "#", label: "Contact" },
];

export default function Navbar() {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <nav className="fixed top-0 left-0 z-10 w-full border-b border-gray-100 bg-white font-['Open_Sans'] shadow-sm">
      <div className="mx-auto max-w-[1344px] px-4 sm:px-6 lg:px-8">
        <div className="flex h-[58px] justify-between">
          <div className="flex">
            <div className="flex shrink-0 items-center">
              <a
                href="/"
                className="text-lg font-bold text-[#4a4a4a] hover:text-[#363636]"
                id="siteLogoText"
              >
                Code Self Study
              </a>
            </div>
          </div>
          <div className="hidden sm:ml-6 sm:flex sm:space-x-8">
            {LINKS.map((link) => (
              <a
                key={link.label}
                href={link.href}
                className="inline-flex items-center px-[13.5px] py-[9px] text-[1rem] leading-6 font-medium text-[#4a4a4a] hover:bg-gray-50 hover:text-[#363636]"
              >
                {link.label}
              </a>
            ))}
          </div>
          <div className="-mr-2 flex items-center sm:hidden">
            <button
              onClick={() => setIsOpen(!isOpen)}
              type="button"
              className="inline-flex items-center justify-center rounded-md p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-500 focus:ring-2 focus:ring-indigo-500 focus:outline-none focus:ring-inset"
              aria-controls="mobile-menu"
              aria-expanded="false"
            >
              <span className="sr-only">Open main menu</span>
              {!isOpen ? (
                <svg
                  className="block h-6 w-6"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  aria-hidden="true"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M4 6h16M4 12h16M4 18h16"
                  />
                </svg>
              ) : (
                <svg
                  className="block h-6 w-6"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  aria-hidden="true"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              )}
            </button>
          </div>
        </div>
      </div>

      {isOpen && (
        <div className="sm:hidden" id="mobile-menu">
          <div className="space-y-1 pt-2 pb-3">
            {LINKS.map((link) => (
              <a
                key={link.label}
                href={link.href}
                className="block border-l-4 border-transparent py-2 pr-4 pl-3 text-base font-medium text-[#4a4a4a] hover:border-gray-300 hover:bg-gray-50 hover:text-[#363636]"
              >
                {link.label}
              </a>
            ))}
          </div>
        </div>
      )}
    </nav>
  );
}
