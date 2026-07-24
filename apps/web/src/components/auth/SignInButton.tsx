import { useEffect, useState } from "react";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import UserMenu from "@/components/auth/UserMenu";
import type { Account } from "@/components/auth/UserMenu";

// Sign-in / sign-out control for the navbar.
//
// Auth now lives in a first-party, HTTP-only session cookie that the Go server
// sets during the WorkOS hosted flow (see apps/api/internal/session). The
// browser holds no WorkOS tokens, so this component just asks the cookie-gated
// /api/me who the user is: 200 -> signed in (render avatar + email + Sign Out),
// 401 -> signed out (render Sign In). Signing in and out are plain navigations
// to the server's /auth routes, which run the hosted flow and set or clear the
// cookie, then send the user back to where they were.
export default function SignInButton() {
  const [status, setStatus] = useState<"loading" | "in" | "out">("loading");
  const [account, setAccount] = useState<Account | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetch("/api/me", {
      credentials: "same-origin",
      headers: { Accept: "application/json" },
    })
      .then((res) =>
        res.ok ? (res.json() as Promise<Partial<Account>>) : null
      )
      .then((data) => {
        if (cancelled) return;
        if (data) {
          setAccount({
            email: data.email ?? "",
            username: data.username ?? "",
            avatar: data.avatar ?? "",
          });
          setStatus("in");
        } else {
          setStatus("out");
        }
      })
      .catch(() => {
        // Network error — treat as signed out rather than crashing the navbar.
        if (!cancelled) setStatus("out");
      });
    return () => {
      cancelled = true;
    };
  }, []);

  if (status === "in" && account) {
    // The username is shown here, never the email — email on a public page is a
    // privacy leak (#351). UserMenu falls back to email only when the server
    // runs without a database and /api/me returns no username.
    return (
      <UserMenu
        account={account}
        onSignOut={() => navigateAuth("/auth/logout")}
      />
    );
  }

  return (
    <button
      type="button"
      disabled={status === "loading"}
      onClick={() => navigateAuth("/auth/login")}
      className={cn(buttonVariants({ variant: "default", size: "sm" }))}
    >
      Sign In
    </button>
  );
}

// navigateAuth sends the browser to a server /auth route, passing the current
// path as returnTo so the server brings the user back here afterwards. The
// server validates returnTo to a same-origin path, so a crafted value can't
// turn this into an open redirect.
function navigateAuth(path: string) {
  // Keep the full in-app location (query + hash), not just the path, so filters
  // and anchors survive the /auth round-trip. The server re-validates this to a
  // same-origin path (see session.safeReturnTo).
  const returnTo =
    window.location.pathname + window.location.search + window.location.hash;
  window.location.assign(`${path}?returnTo=${encodeURIComponent(returnTo)}`);
}
