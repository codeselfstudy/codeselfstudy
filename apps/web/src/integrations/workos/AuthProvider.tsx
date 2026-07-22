import { AuthKitProvider } from "@workos-inc/authkit-react";
import type { ReactNode } from "react";

// safeReturnTo resolves a post-sign-in `returnTo` against the current origin and
// returns a same-origin relative target, or null. `state` round-trips through
// the WorkOS hosted flow and could be attacker-influenced, so an absolute or
// protocol-relative external URL (an open-redirect attempt) is rejected, as is a
// malformed value. Exported for direct testing.
export function safeReturnTo(
  returnTo: string | null | undefined,
  origin: string
): string | null {
  if (!returnTo) return null;
  try {
    const url = new URL(returnTo, origin);
    if (url.origin === origin) {
      return url.pathname + url.search + url.hash;
    }
  } catch {
    // Malformed returnTo — ignore.
  }
  return null;
}

// Client-side WorkOS AuthKit provider, ported from the doolin fork's
// TanStack-Start version. AuthKit-react is a browser-only SPA integration (PKCE,
// tokens held in the browser), so this has no server-rendered form and must be
// mounted inside a `client:only="react"` island.
//
// After the hosted sign-in flow redirects back, `onRedirectCallback` returns the
// user to the page they started from — the `returnTo` the sign-in call stashed
// on `state`. The fork navigated with TanStack Router's `useNavigate`; on a
// static Astro site there is no client router, so a full-page
// `window.location.assign` is the equivalent (origin-guarded by safeReturnTo).
export default function AuthProvider({ children }: { children: ReactNode }) {
  return (
    <AuthKitProvider
      clientId={import.meta.env.VITE_WORKOS_CLIENT_ID}
      apiHostname={import.meta.env.VITE_WORKOS_API_HOSTNAME}
      onRedirectCallback={({ state }) => {
        const target = safeReturnTo(
          (state as { returnTo?: string } | null)?.returnTo,
          window.location.origin
        );
        if (target) window.location.assign(target);
      }}
    >
      {children}
    </AuthKitProvider>
  );
}
