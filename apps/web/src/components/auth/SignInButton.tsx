import { useEffect, useState } from "react";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import UserMenu from "@/components/auth/UserMenu";
import type { Account } from "@/components/auth/UserMenu";
import {
  clearAccountHint,
  isSignedInHint,
  readAccountHint,
  writeAccountHint,
} from "@/lib/authHint";

// Sign-in / sign-out control for the navbar.
//
// Auth lives in a first-party, HTTP-only session cookie that the Go server sets
// during the WorkOS hosted flow (see apps/api/internal/session). The browser
// holds no WorkOS tokens, so this component asks the cookie-gated /api/me who
// the user is: 200 -> signed in, 401 -> signed out. Signing in and out are plain
// navigations to the server's /auth routes, which run the hosted flow and set or
// clear the cookie, then send the user back to where they were.
//
// That fetch takes a round-trip, and rendering Sign In while it is in flight
// flashed the wrong control at every signed-in visitor (#367). So the first
// render is seeded from the auth hint instead — a server-managed flag cookie
// plus a cached username/avatar (see lib/authHint). The hint decides what paints
// first; /api/me remains the only thing that decides what is true.
type AuthState = { status: "in"; account: Account } | { status: "out" };

// initialState runs once, synchronously, before the first paint.
function initialState(): AuthState {
  if (!isSignedInHint()) return { status: "out" };
  // Signed in per the server, but the cached detail may be missing (first visit
  // after signing in on another device, cleared storage). UserMenu renders a
  // generic "Account" label for an empty profile, so this still paints a real
  // control rather than a flash.
  const hint = readAccountHint();
  return {
    status: "in",
    account: {
      email: "",
      username: hint?.username ?? "",
      avatar: hint?.avatar ?? "",
    },
  };
}

export default function SignInButton() {
  const [state, setState] = useState<AuthState>(initialState);

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
          const account: Account = {
            email: data.email ?? "",
            username: data.username ?? "",
            avatar: data.avatar ?? "",
          };
          setState({ status: "in", account });
          writeAccountHint({
            username: account.username,
            avatar: account.avatar,
          });
        } else {
          // The hint said signed in and the server disagrees, or there was no
          // hint at all. Either way the cache is now wrong.
          clearAccountHint();
          setState({ status: "out" });
        }
      })
      .catch(() => {
        // Network error. Leave both the state and the cache exactly as seeded:
        // a fetch that never arrived is not the server saying "signed out", and
        // flipping to Sign In here would re-create the very flash this fixes,
        // triggered by flakiness instead of latency. Only a definitive /api/me
        // answer changes what is displayed. Someone with no hint was already
        // showing Sign In, so they see no change either.
      });
    return () => {
      cancelled = true;
    };
  }, []);

  if (state.status === "in") {
    // The username is shown here, never the email — email on a public page is a
    // privacy leak (#351). UserMenu falls back to email only when the server
    // runs without a database and /api/me returns no username.
    return (
      <UserMenu
        account={state.account}
        onSignOut={() => {
          // The server clears the flag cookie on /auth/logout; the cached
          // username is ours to drop.
          clearAccountHint();
          navigateAuth("/auth/logout");
        }}
      />
    );
  }

  return (
    <button
      type="button"
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
