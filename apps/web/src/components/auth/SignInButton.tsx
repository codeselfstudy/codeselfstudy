import { useEffect, useState } from "react";
import { useAuth } from "@workos-inc/authkit-react";

import { apiFetch } from "@/lib/api";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

// Sign-in / sign-out control, ported from the doolin fork's workos-user.tsx and
// restyled with the shared shadcn button. Signed out: a "Sign In" button that
// launches the WorkOS hosted flow, passing the current path as `returnTo` so the
// AuthProvider's redirect callback can bring the user back here. Signed in: the
// user (email from the server, name as a fallback) with a "Sign Out" button.
export default function SignInButton() {
  const { user, isLoading, signIn, signOut, getAccessToken } = useAuth();
  const [email, setEmail] = useState<string | null>(null);

  // Prove the auth loop end to end: with a session, call the WorkOS-gated
  // /api/me with the access token and show the email the *server* returns (not
  // the client-side user object), confirming the Go middleware accepted the
  // token. Any failure just leaves email null and the name shows instead.
  useEffect(() => {
    if (!user) {
      setEmail(null);
      return;
    }
    let cancelled = false;
    apiFetch("/api/me", getAccessToken)
      .then((res) => (res.ok ? res.json() : null))
      .then((data: { email?: string } | null) => {
        if (!cancelled) setEmail(data?.email ?? null);
      })
      .catch(() => {
        if (!cancelled) setEmail(null);
      });
    return () => {
      cancelled = true;
    };
  }, [user, getAccessToken]);

  if (user) {
    const label =
      email ?? [user.firstName, user.lastName].filter(Boolean).join(" ");
    return (
      <div className="flex items-center gap-2">
        {user.profilePictureUrl && (
          <img
            src={user.profilePictureUrl}
            alt=""
            className="h-7 w-7 rounded-full"
          />
        )}
        {label && (
          <span className="hidden text-sm font-medium text-[#4a4a4a] sm:inline">
            {label}
          </span>
        )}
        <button
          type="button"
          onClick={() => signOut()}
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
      onClick={() => signIn({ state: { returnTo: window.location.pathname } })}
      disabled={isLoading}
      className={cn(buttonVariants({ variant: "default", size: "sm" }))}
    >
      Sign In
    </button>
  );
}
