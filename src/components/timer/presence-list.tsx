import { Crown, Users } from "lucide-react";
import type { PresenceState } from "@/lib/timer/ws-client";
import { cn } from "@/lib/utils";

interface PresenceListProps {
  presence: PresenceState | null;
  currentUserId?: string;
  className?: string;
}

export function PresenceList({
  presence,
  currentUserId,
  className,
}: PresenceListProps) {
  if (!presence) {
    return null;
  }

  return (
    <div className={cn("rounded-lg border p-4", className)}>
      <div className="mb-3 flex items-center gap-2 text-sm font-medium">
        <Users className="h-4 w-4" />
        Participants ({presence.count})
      </div>

      <ul className="space-y-2">
        {presence.users.map((user) => (
          <li
            key={user.id}
            className={cn(
              "flex items-center gap-2 text-sm",
              user.id === currentUserId && "font-medium"
            )}
          >
            {user.isAdmin && (
              <Crown className="h-4 w-4 text-yellow-500" aria-label="Admin" />
            )}
            <span>{user.name}</span>
            {user.id === currentUserId && (
              <span className="text-muted-foreground">(you)</span>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}
