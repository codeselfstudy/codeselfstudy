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
import SignInButton from "@/components/auth/SignInButton";

const LINKS = [
  // { href: "/", label: "Home" },
  { href: "/events/", label: "Events" },
  { href: "/about/", label: "About" },
  // { href: "/contact/", label: "Contact" },
];

export default function Navbar() {
  const [isOpen, setIsOpen] = React.useState(false);

  // This island is mounted with client:only="react" (see Layout.astro): the
  // sign-in control fetches the cookie-gated /api/me on mount, which only works
  // in the browser. The theme toggle and sign-in control are single,
  // always-visible instances — the mobile drawer holds only the nav links — so
  // exactly one SignInButton is mounted (no duplicate /api/me). The theme itself
  // is set before paint by the inline script in Layout.astro, so dark mode still
  // works even though the nav hydrates late.
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
            <ThemeToggle />
            <SignInButton />
            <div className="sm:hidden">
              <Sheet open={isOpen} onOpenChange={setIsOpen}>
                <SheetTrigger
                  className={buttonVariants({
                    variant: "ghost",
                    size: "icon",
                  })}
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
