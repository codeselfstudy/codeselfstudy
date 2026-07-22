import { useAuth } from "@workos-inc/authkit-react";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

// Sign-in / sign-out control, ported from the doolin fork's workos-user.tsx and
// restyled with the shared shadcn button. Signed out: a "Sign In" button that
// launches the WorkOS hosted flow, passing the current path as `returnTo` so the
// AuthProvider's redirect callback can bring the user back here. Signed in: the
// user's name (avatar when present) with a "Sign Out" button.
export default function SignInButton() {
  const { user, isLoading, signIn, signOut } = useAuth();

  if (user) {
    const name = [user.firstName, user.lastName].filter(Boolean).join(" ");
    return (
      <div className="flex items-center gap-2">
        {user.profilePictureUrl && (
          <img
            src={user.profilePictureUrl}
            alt=""
            className="h-7 w-7 rounded-full"
          />
        )}
        {name && (
          <span className="hidden text-sm font-medium text-[#4a4a4a] sm:inline">
            {name}
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
