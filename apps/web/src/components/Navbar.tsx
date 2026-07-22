import * as React from "react";
import { Menu } from "lucide-react";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import {
  Sheet,
  SheetContent,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import ThemeToggle from "@/components/ThemeToggle";

const LINKS = [
  { href: "/", label: "Home" },
  { href: "/about/", label: "About" },
  { href: "/events/", label: "Events" },
  // { href: "/contact/", label: "Contact" },
];

export default function Navbar() {
  const [isOpen, setIsOpen] = React.useState(false);

  return (
    <nav className="bg-background border-border fixed top-0 left-0 z-10 w-full border-b shadow-sm">
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex h-[58px] items-center justify-between">
          <div className="flex shrink-0 items-center">
            <a
              href="/"
              className="text-foreground text-lg font-bold hover:no-underline hover:opacity-80"
              id="siteLogoText"
            >
              Code Self Study
            </a>
          </div>
          <div className="flex items-center gap-1">
            <ThemeToggle />
            <div className="hidden sm:flex sm:items-center sm:space-x-4">
              {LINKS.map((link) => (
                <a
                  key={link.label}
                  href={link.href}
                  className={cn(
                    buttonVariants({ variant: "ghost" }),
                    "text-foreground text-[1rem] font-medium"
                  )}
                >
                  {link.label}
                </a>
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
                      <a
                        key={link.label}
                        href={link.href}
                        className="text-foreground text-lg font-medium hover:no-underline hover:opacity-80"
                        onClick={() => setIsOpen(false)}
                      >
                        {link.label}
                      </a>
                    ))}
                  </div>
                </SheetContent>
              </Sheet>
            </div>
          </div>
        </div>
      </div>
    </nav>
  );
}
