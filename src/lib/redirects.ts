import type { RedirectConfig } from "@/data/redirects";
import { redirects } from "@/data/redirects";

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
      if (
        pathname.startsWith(prefix + "/") &&
        pathname.length > prefix.length + 1
      ) {
        return config;
      }
    }
  }

  return null;
}

// Re-export types for convenience
export type { RedirectConfig } from "@/data/redirects";
