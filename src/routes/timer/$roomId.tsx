import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useCallback, useEffect, useState } from "react";
import { AlertCircle, Share2, Volume2, VolumeX } from "lucide-react";
import { useAuth } from "@workos-inc/authkit-react";
import { useTimerWS } from "@/lib/timer/ws-client";
import { generateGuestName } from "@/lib/timer/names";
import { TimerDisplay } from "@/components/timer/timer-display";
import { TimerControls } from "@/components/timer/timer-controls";
import { PresenceList } from "@/components/timer/presence-list";
import { Button } from "@/components/ui/button";
import { createMetadata } from "@/lib/metadata";
import { SignInButton } from "@/components/workos-user";
import { PageWrapper } from "@/components/page-wrapper";

export const Route = createFileRoute("/timer/$roomId")({
  component: TimerRoom,
  head: ({ params }) => ({
    meta: createMetadata({
      title: `Timer Room: ${params.roomId}`,
      description: "Join this Pomodoro timer session.",
    }),
  }),
});

function TimerRoom() {
  const { roomId } = Route.useParams();
  const { user, isLoading: isAuthLoading } = useAuth();
  const navigate = useNavigate();

  const [token, setToken] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isMuted, setIsMuted] = useState(() => {
    if (typeof window !== "undefined") {
      return localStorage.getItem("timer-muted") === "true";
    }
    return false;
  });
  const [guestName, setGuestName] = useState<string | null>(null);
  const [sessionId, setSessionId] = useState<string | null>(null);

  // Generate or retrieve session ID for guests
  useEffect(() => {
    if (typeof window === "undefined") return;
    // Wait for auth to finish loading before generating guest name
    if (isAuthLoading) return;

    let id = sessionStorage.getItem("timer-session-id");
    if (!id) {
      id = crypto.randomUUID();
      sessionStorage.setItem("timer-session-id", id);
    }
    setSessionId(id);

    if (!user) {
      setGuestName(generateGuestName(id));
    }
  }, [user, isAuthLoading]);

  // Fetch room info and token
  useEffect(() => {
    // Wait for auth to finish loading
    if (isAuthLoading) return;

    async function fetchRoom() {
      setIsLoading(true);
      setError(null);

      try {
        const userId = user?.id || sessionId;
        if (!userId) return;

        const response = await fetch(
          `/api/timer/rooms/${encodeURIComponent(roomId)}?userId=${encodeURIComponent(userId)}`
        );

        if (!response.ok) {
          if (response.status === 404) {
            setError("Room not found");
          } else {
            const data = await response.json();
            setError(data.message || "Failed to load room");
          }
          return;
        }

        const data = await response.json();
        setToken(data.token);
      } catch (err) {
        setError("Failed to connect to room");
      } finally {
        setIsLoading(false);
      }
    }

    if (sessionId || user) {
      fetchRoom();
    }
  }, [roomId, user, sessionId, isAuthLoading]);

  // WebSocket connection
  const handleError = useCallback((code: string, message: string) => {
    console.error(`Timer error [${code}]: ${message}`);
    if (code === "ROOM_NOT_FOUND") {
      setError("Room not found");
    }
  }, []);

  const handleSessionEnded = useCallback(
    (reason: string) => {
      alert(`Session ended: ${reason}`);
      navigate({ to: "/timer" });
    },
    [navigate]
  );

  const ws = useTimerWS({
    roomId,
    token: token || "",
    name: user?.firstName || guestName || undefined,
    onError: handleError,
    onSessionEnded: handleSessionEnded,
  });

  // Only connect when we have a token
  const isConnected = token && ws.status === "connected";

  // Toggle mute
  const toggleMute = () => {
    const newMuted = !isMuted;
    setIsMuted(newMuted);
    if (typeof window !== "undefined") {
      localStorage.setItem("timer-muted", String(newMuted));
    }
  };

  // Share room
  const shareRoom = async () => {
    const url = window.location.href;
    // Check if Web Share API is available (not all browsers support it)
    if (typeof navigator.share === "function") {
      await navigator.share({
        title: "Join my Pomodoro timer",
        url,
      });
    } else {
      await navigator.clipboard.writeText(url);
      alert("Link copied to clipboard!");
    }
  };

  if (isLoading) {
    return (
      <PageWrapper>
        <div className="flex min-h-[60vh] items-center justify-center">
          <div className="text-muted-foreground">Loading room...</div>
        </div>
      </PageWrapper>
    );
  }

  if (error) {
    return (
      <PageWrapper>
        <div className="flex min-h-[60vh] flex-col items-center justify-center gap-4">
          <AlertCircle className="text-destructive h-12 w-12" />
          <h2 className="text-xl font-semibold">{error}</h2>
          <Button onClick={() => navigate({ to: "/timer" })}>
            Back to Timer Home
          </Button>
        </div>
      </PageWrapper>
    );
  }

  return (
    <PageWrapper>
      <div className="container mx-auto max-w-4xl px-4 py-8">
        {/* Header */}
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold">Room: {roomId}</h1>
            <div className="text-muted-foreground text-sm">
              {ws.status === "connected"
                ? "Connected"
                : ws.status === "connecting"
                  ? "Connecting..."
                  : "Disconnected"}
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Button variant="ghost" size="icon" onClick={toggleMute}>
              {isMuted ? (
                <VolumeX className="h-5 w-5" />
              ) : (
                <Volume2 className="h-5 w-5" />
              )}
            </Button>
            <Button variant="ghost" size="icon" onClick={shareRoom}>
              <Share2 className="h-5 w-5" />
            </Button>
          </div>
        </div>

        {/* Main content */}
        <div className="grid gap-6 lg:grid-cols-[1fr_280px]">
          {/* Timer area */}
          <div className="space-y-6">
            {ws.timerState && (
              <TimerDisplay state={ws.timerState} className="min-h-[300px]" />
            )}

            {ws.timerState && (
              <TimerControls
                state={ws.timerState}
                isAdmin={ws.isAdmin}
                onStart={ws.start}
                onPause={ws.pause}
                onReset={ws.reset}
                onSkip={ws.skip}
                onEndSession={ws.endSession}
              />
            )}
          </div>

          {/* Sidebar */}
          <div className="space-y-4">
            <PresenceList
              presence={ws.presence}
              currentUserId={user?.id || sessionId || undefined}
            />

            {!user && (
              <div className="bg-muted rounded-lg p-4 text-center text-sm">
                <p className="mb-2">
                  You're viewing as <strong>{guestName}</strong>
                </p>
                <SignInButton />
              </div>
            )}
          </div>
        </div>
      </div>
    </PageWrapper>
  );
}
