import { useEffect, useState } from "react";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type Account = { email: string; name: string; avatar: string };

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
            name: data.name ?? "",
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
    const label = account.email || account.name;
    return (
      <div className="flex items-center gap-2">
        {account.avatar && (
          <img src={account.avatar} alt="" className="h-7 w-7 rounded-full" />
        )}
        {label && (
          <span className="text-foreground hidden text-sm font-medium sm:inline">
            {label}
          </span>
        )}
        <button
          type="button"
          onClick={() => navigateAuth("/auth/logout")}
          className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
        >
          Sign Out
        </button>
      </div>
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
  const returnTo = window.location.pathname;
  window.location.assign(`${path}?returnTo=${encodeURIComponent(returnTo)}`);
}
