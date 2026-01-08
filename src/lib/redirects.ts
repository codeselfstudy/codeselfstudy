import type { Redirect, RedirectConfig } from "@/data/redirects";
import { redirects } from "@/data/redirects";

// Re-export types for convenience
export type { RedirectConfig } from "@/data/redirects";

// Pre-compute redirects for optimal performance
const exactRedirects: Redirect = {};
const wildcardRedirects: Array<{ prefix: string; config: RedirectConfig }> = [];

for (const [pattern, config] of Object.entries(redirects)) {
  if (pattern.endsWith("/*")) {
    wildcardRedirects.push({
      prefix: pattern.slice(0, -2),
      config,
    });
  } else {
    exactRedirects[pattern] = config;
  }
}

/**
 * Find a redirect for the given pathname.
 * Supports both exact matches and wildcard patterns (e.g., "/blog/*").
 * Optimized with pre-computed exact and wildcard lookups.
 *
 * @param pathname - The pathname to match
 * @returns The redirect configuration if found, null otherwise
 */
export function findRedirect(pathname: string): RedirectConfig | null {
  // O(1) exact match first for performance
  if (exactRedirects[pathname]) {
    return exactRedirects[pathname];
  }

  // Only check wildcards if exact match fails
  for (const { prefix, config } of wildcardRedirects) {
    if (
      pathname.startsWith(prefix + "/") &&
      pathname.length > prefix.length + 1
    ) {
      return config;
    }
  }

  return null;
}
