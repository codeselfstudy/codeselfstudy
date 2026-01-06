import * as React from "react";
import { Link } from "@tanstack/react-router";
import { Menu } from "lucide-react";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import {
  Sheet,
  SheetContent,
  SheetTrigger,
  SheetTitle,
} from "@/components/ui/sheet";

const LINKS = [
  { href: "/", label: "Home" },
  { href: "/about", label: "About" },
  { href: "/events", label: "Events" },
  // { href: "/contact", label: "Contact" },
];

export default function Navbar() {
  const [isOpen, setIsOpen] = React.useState(false);

  return (
    <nav className="fixed top-0 left-0 z-10 w-full border-b border-gray-100 bg-white shadow-sm">
      <div className="mx-auto max-w-[1344px] px-4 sm:px-6 lg:px-8">
        <div className="flex h-[58px] items-center justify-between">
          <div className="flex shrink-0 items-center">
            <Link
              to="/"
              className="text-lg font-bold text-[#4a4a4a] hover:text-[#363636]"
              id="siteLogoText"
            >
              Code Self Study
            </Link>
          </div>
          <div className="hidden sm:flex sm:space-x-4">
            {LINKS.map((link) => (
              <Link
                key={link.label}
                to={link.href}
                className={cn(
                  buttonVariants({ variant: "ghost" }),
                  "text-[1rem] font-medium text-[#4a4a4a]"
                )}
              >
                {link.label}
              </Link>
            ))}
          </div>
          <div className="sm:hidden">
            <Sheet open={isOpen} onOpenChange={setIsOpen}>
              <SheetTrigger
                className={buttonVariants({ variant: "ghost", size: "icon" })}
              >
                <Menu className="h-6 w-6" />
                <span className="sr-only">Open main menu</span>
              </SheetTrigger>
              <SheetContent side="right">
                <SheetTitle>Menu</SheetTitle>
                <div className="mt-6 flex flex-col space-y-4">
                  {LINKS.map((link) => (
                    <Link
                      key={link.label}
                      to={link.href}
                      className="text-lg font-medium text-[#4a4a4a] hover:text-[#363636]"
                      onClick={() => setIsOpen(false)}
                    >
                      {link.label}
                    </Link>
                  ))}
                </div>
              </SheetContent>
            </Sheet>
          </div>
        </div>
      </div>
    </nav>
  );
}
