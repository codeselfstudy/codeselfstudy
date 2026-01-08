import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { Clock, Users } from "lucide-react";
import { useAuth } from "@workos-inc/authkit-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { createMetadata } from "@/lib/metadata";
import { SignInButton } from "@/components/workos-user";
import { PageWrapper } from "@/components/page-wrapper";

export const Route = createFileRoute("/timer/")({
  component: TimerLanding,
  head: () => ({
    meta: createMetadata({
      title: "Group Pomodoro Timer",
      description:
        "Stay focused together with synchronized Pomodoro timers. Create a room and invite others to join.",
    }),
  }),
});

function TimerLanding() {
  const { user, isLoading: isAuthLoading } = useAuth();
  const navigate = useNavigate();

  const [slug, setSlug] = useState("");
  const [focusMinutes, setFocusMinutes] = useState(25);
  const [breakMinutes, setBreakMinutes] = useState(5);
  const [longBreakMinutes, setLongBreakMinutes] = useState(15);
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [joinRoomId, setJoinRoomId] = useState("");

  // Show loading state while auth is initializing
  const isAuthenticated = !isAuthLoading && user;

  const handleCreateRoom = async () => {
    if (!user) {
      setError("You must be signed in to create a room");
      return;
    }

    setIsCreating(true);
    setError(null);

    try {
      const response = await fetch("/api/timer/rooms", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          userId: user.id,
          slug: slug || undefined,
          focusDuration: focusMinutes * 60 * 1000,
          breakDuration: breakMinutes * 60 * 1000,
          longBreakDuration: longBreakMinutes * 60 * 1000,
        }),
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.message || "Failed to create room");
      }

      const { roomId } = await response.json();
      navigate({ to: "/timer/$roomId", params: { roomId } });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create room");
    } finally {
      setIsCreating(false);
    }
  };

  const handleJoinRoom = () => {
    if (joinRoomId.trim()) {
      navigate({ to: "/timer/$roomId", params: { roomId: joinRoomId.trim() } });
    }
  };

  return (
    <PageWrapper>
      <div className="container mx-auto max-w-4xl px-4 py-8">
        <div className="mb-8 text-center">
          <h1 className="text-4xl font-bold">Group Pomodoro Timer</h1>
          <p className="text-muted-foreground mt-2">
            Stay focused together with synchronized timers
          </p>
        </div>

        <div className="grid gap-6 md:grid-cols-2">
          {/* Create Room */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Clock className="h-5 w-5" />
                Create a Room
              </CardTitle>
              <CardDescription>
                {isAuthenticated
                  ? "Start a new timer session and invite others"
                  : "Sign in to create a timer room"}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {isAuthLoading ? (
                <div className="text-muted-foreground py-4 text-center">
                  Loading...
                </div>
              ) : isAuthenticated ? (
                <>
                  <div className="space-y-2">
                    <Label htmlFor="slug">Custom URL (optional)</Label>
                    <div className="flex items-center gap-2">
                      <span className="text-muted-foreground text-sm">
                        /timer/
                      </span>
                      <Input
                        id="slug"
                        placeholder="team-standup"
                        value={slug}
                        onChange={(e) => setSlug(e.target.value)}
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-3 gap-3">
                    <div className="space-y-2">
                      <Label htmlFor="focus">Focus (min)</Label>
                      <Input
                        id="focus"
                        type="number"
                        min={1}
                        max={120}
                        value={focusMinutes}
                        onChange={(e) =>
                          setFocusMinutes(Number(e.target.value))
                        }
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="break">Break (min)</Label>
                      <Input
                        id="break"
                        type="number"
                        min={1}
                        max={60}
                        value={breakMinutes}
                        onChange={(e) =>
                          setBreakMinutes(Number(e.target.value))
                        }
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="longBreak">Long (min)</Label>
                      <Input
                        id="longBreak"
                        type="number"
                        min={1}
                        max={60}
                        value={longBreakMinutes}
                        onChange={(e) =>
                          setLongBreakMinutes(Number(e.target.value))
                        }
                      />
                    </div>
                  </div>

                  {error && <p className="text-destructive text-sm">{error}</p>}

                  <Button
                    onClick={handleCreateRoom}
                    disabled={isCreating}
                    className="w-full"
                  >
                    {isCreating ? "Creating..." : "Create Room"}
                  </Button>
                </>
              ) : (
                <SignInButton />
              )}
            </CardContent>
          </Card>

          {/* Join Room */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Users className="h-5 w-5" />
                Join a Room
              </CardTitle>
              <CardDescription>
                Enter a room code or custom URL to join an existing session
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="joinRoomId">Room Code or URL</Label>
                <Input
                  id="joinRoomId"
                  placeholder="abc123 or team-standup"
                  value={joinRoomId}
                  onChange={(e) => setJoinRoomId(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") handleJoinRoom();
                  }}
                />
              </div>

              <Button
                onClick={handleJoinRoom}
                disabled={!joinRoomId.trim()}
                variant="secondary"
                className="w-full"
              >
                Join Room
              </Button>
            </CardContent>
          </Card>
        </div>

        {/* Feature highlights */}
        <div className="mt-12 grid gap-4 text-center md:grid-cols-3">
          <div>
            <div className="bg-primary/10 mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full">
              <Clock className="text-primary h-6 w-6" />
            </div>
            <h3 className="font-semibold">Pomodoro Technique</h3>
            <p className="text-muted-foreground text-sm">
              25 min focus, 5 min break, 15 min long break
            </p>
          </div>
          <div>
            <div className="bg-primary/10 mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full">
              <Users className="text-primary h-6 w-6" />
            </div>
            <h3 className="font-semibold">Real-time Sync</h3>
            <p className="text-muted-foreground text-sm">
              Everyone sees the same timer, synced in real-time
            </p>
          </div>
          <div>
            <div className="bg-primary/10 mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full">
              <Clock className="text-primary h-6 w-6" />
            </div>
            <h3 className="font-semibold">Custom Durations</h3>
            <p className="text-muted-foreground text-sm">
              Customize focus and break durations to fit your workflow
            </p>
          </div>
        </div>
      </div>
    </PageWrapper>
  );
}
