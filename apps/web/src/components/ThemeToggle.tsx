import * as React from "react";
import { Moon, Sun } from "lucide-react";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

/**
 * Light/dark toggle. The inline script in Layout.astro sets the initial `.dark`
 * class before paint (device setting by default, or the user's stored choice).
 * This button flips that class and persists the choice under the "theme" key.
 * Icons are shown via the `dark:` variant (CSS), so they never flash on hydration.
 */
export default function ThemeToggle() {
  const [isDark, setIsDark] = React.useState(false);

  React.useEffect(() => {
    setIsDark(document.documentElement.classList.contains("dark"));

    // Track the OS setting, but only while the user hasn't chosen explicitly.
    const mq = window.matchMedia?.("(prefers-color-scheme: dark)");
    if (!mq) return;
    const onChange = (event: MediaQueryListEvent) => {
      let stored: string | null = null;
      try {
        stored = localStorage.getItem("theme");
      } catch {
        /* ignore */
      }
      if (!stored) {
        document.documentElement.classList.toggle("dark", event.matches);
        setIsDark(event.matches);
      }
    };
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);

  const toggle = () => {
    const next = !document.documentElement.classList.contains("dark");
    document.documentElement.classList.toggle("dark", next);
    try {
      localStorage.setItem("theme", next ? "dark" : "light");
    } catch {
      /* ignore */
    }
    setIsDark(next);
  };

  return (
    <button
      type="button"
      onClick={toggle}
      aria-label="Toggle dark mode"
      aria-pressed={isDark}
      className={cn(
        buttonVariants({ variant: "ghost", size: "icon" }),
        "text-foreground"
      )}
    >
      <Moon className="h-5 w-5 dark:hidden" />
      <Sun className="hidden h-5 w-5 dark:block" />
      <span className="sr-only">Toggle dark mode</span>
    </button>
  );
}
