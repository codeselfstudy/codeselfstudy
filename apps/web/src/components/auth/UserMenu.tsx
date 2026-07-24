import { buttonVariants } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuLinkItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";

// Account is the shape SignInButton reads from /api/me and hands to the menu.
// The label prefers the username (public-safe) and only falls back to the email
// when the server runs without a database and returns no username.
export type Account = { email: string; username: string; avatar: string };

// UserMenu is the signed-in navbar control: an avatar + username trigger that
// opens a dropdown to the settings page and sign-out. SignInButton still owns
// the single /api/me fetch and passes the resolved account down, so the
// one-request-per-load invariant (see Navbar.tsx) is unchanged. Keyboard and
// screen-reader behaviour come from the Menu primitive.
export default function UserMenu({
  account,
  onSignOut,
}: {
  account: Account;
  onSignOut: () => void;
}) {
  const label = account.username || account.email;
  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        aria-label={label ? `Account menu for ${label}` : "Account menu"}
        className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
      >
        {account.avatar && (
          <img src={account.avatar} alt="" className="h-6 w-6 rounded-full" />
        )}
        {/* Always visible so the trigger stays a real affordance even with no
            avatar (e.g. on mobile); falls back to "Account" for an empty
            profile. */}
        <span className="max-w-[16ch] truncate">{label || "Account"}</span>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        {label && <DropdownMenuLabel>{label}</DropdownMenuLabel>}
        <DropdownMenuLinkItem href="/settings/">Settings</DropdownMenuLinkItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={onSignOut}>Sign Out</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
