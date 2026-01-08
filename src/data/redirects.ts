/**
 * This site contains redirects.
 */
type RedirectConfig = { to: string; status: 308 | 307 };
type Redirect = Record<string, RedirectConfig>;

const redirects: Redirect = {
  "/book": { to: "/learn", status: 308 },
  "/book/": { to: "/learn", status: 308 },
  "/apps": { to: "/learn", status: 308 },
  "/apps/": { to: "/learn", status: 308 },
  "/autism": { to: "/learn", status: 308 },
  "/autism/": { to: "/learn", status: 308 },
  "/contribute": { to: "/learn", status: 308 },
  "/contribute/": { to: "/learn", status: 308 },
  "/sponsors": { to: "/credits", status: 308 },
  "/sponsors/": { to: "/credits", status: 308 },
  "/parking": { to: "/events", status: 308 },
  "/parking/": { to: "/events", status: 308 },
  "/school": { to: "/learn", status: 308 },
  "/school/": { to: "/learn", status: 308 },
  "/support-free-software": { to: "/learn", status: 308 },
  "/support-free-software/": { to: "/learn", status: 308 },
  "/tutorials": { to: "/learn", status: 308 },
  "/tutorials/": { to: "/learn", status: 308 },
  "/featured": { to: "/learn", status: 308 },
  "/featured/": { to: "/learn", status: 308 },
  "/hire-programmers": { to: "/jobs", status: 308 },
  "/hire-programmers/": { to: "/jobs", status: 308 },
  "/wiki": { to: "/learn", status: 308 },
  "/wiki/": { to: "/learn", status: 308 },
  "/discounts/algoexpert": { to: "/learn", status: 308 },
  "/discounts/algoexpert/": { to: "/learn", status: 308 },
  "/discounts/digitalocean": { to: "/discounts", status: 308 },
  "/discounts/digitalocean/": { to: "/discounts", status: 308 },
  "/sponsors/digitalocean": { to: "/discounts", status: 308 },
  "/sponsors/digitalocean/": { to: "/discounts", status: 308 },
  "/sponsors/thplibrary": { to: "/discounts", status: 308 },
  "/sponsors/thplibrary/": { to: "/discounts", status: 308 },
  "/home": { to: "/", status: 308 },
  "/home/": { to: "/", status: 308 },
  "/index.html": { to: "/", status: 308 },
  "/support-us/": { to: "/", status: 308 },
  "/support/": { to: "/", status: 308 },
  "/support": { to: "/", status: 308 },
  "/presentations": { to: "/learn", status: 308 },
  "/presentations/": { to: "/learn", status: 308 },
  "/presentations-schedule": { to: "/learn", status: 308 },
  "/blog/*": { to: "/learn", status: 308 },
  "/c": { to: "/forum", status: 308 },
  "/c/": { to: "/forum", status: 308 },
  "/bbs": { to: "/forum", status: 308 },
  "/join": { to: "/events", status: 308 },
  "/welcome": { to: "/events", status: 308 },
  "/checkin": { to: "/forum", status: 308 },
  "/edu": { to: "/learn", status: 308 },
  "/edu/": { to: "/learn", status: 308 },
  "/comment/63": { to: "/", status: 308 },
  "/forms/feedback": { to: "/contact", status: 308 },
  "/wiki/Main_Page": { to: "/contact", status: 308 },
  "/w": { to: "/learn", status: 308 },
  "/w/": { to: "/learn", status: 308 },
  "/wiki/*": { to: "/learn", status: 308 },
  "/programming-notes-wiki/": { to: "/learn", status: 308 },
  "/python-web-scraping-selenium.html": { to: "/learn", status: 308 },
  "/Array": { to: "/learn", status: 308 },
  "/wiki/C%2B%2B": { to: "/learn", status: 308 },
  "/b/*": { to: "/learn", status: 308 },
  "/b/page/6/": { to: "/learn", status: 308 },
  "/b": { to: "/learn", status: 308 },
  "/b/": { to: "/learn", status: 308 },
  "/tools/unicode": { to: "/tools", status: 308 },
  "/tools/unicode/": { to: "/tools", status: 308 },
};

/**
 * Find a redirect for the given pathname.
 * Supports both exact matches and wildcard patterns (e.g., "/blog/*").
 *
 * @param pathname - The pathname to match
 * @returns The redirect configuration if found, null otherwise
 */
export function findRedirect(pathname: string): RedirectConfig | null {
  // First, try exact match for performance
  if (redirects[pathname]) {
    return redirects[pathname];
  }

  // Then try wildcard patterns
  for (const [pattern, config] of Object.entries(redirects)) {
    if (pattern.endsWith("/*")) {
      const prefix = pattern.slice(0, -2); // Remove "/*"
      // Match paths that start with prefix/ and have content after the slash
      if (pathname.startsWith(prefix + "/") && pathname.length > prefix.length + 1) {
        return config;
      }
    }
  }

  return null;
}

export default redirects;
