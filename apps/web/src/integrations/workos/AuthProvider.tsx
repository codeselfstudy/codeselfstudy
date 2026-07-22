import { AuthKitProvider } from "@workos-inc/authkit-react";
import type { ReactNode } from "react";

// Client-side WorkOS AuthKit provider, ported from the doolin fork's
// TanStack-Start version. AuthKit-react is a browser-only SPA integration (PKCE,
// tokens held in the browser), so this has no server-rendered form and must be
// mounted inside a `client:only="react"` island.
//
// After the hosted sign-in flow redirects back, `onRedirectCallback` returns the
// user to the page they started from — the `returnTo` that the sign-in call
// stashed on `state`. The fork navigated with TanStack Router's `useNavigate`;
// on a static Astro site there is no client router, so a full-page
// `window.location.assign` is the right equivalent.
export default function AuthProvider({ children }: { children: ReactNode }) {
  return (
    <AuthKitProvider
      clientId={import.meta.env.VITE_WORKOS_CLIENT_ID}
      apiHostname={import.meta.env.VITE_WORKOS_API_HOSTNAME}
      onRedirectCallback={({ state }) => {
        const returnTo = (state as { returnTo?: string } | null)?.returnTo;
        if (returnTo) {
          window.location.assign(returnTo);
        }
      }}
    >
      {children}
    </AuthKitProvider>
  );
}
