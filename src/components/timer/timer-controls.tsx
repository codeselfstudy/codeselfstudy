import { Pause, Play, RotateCcw, SkipForward, Square } from "lucide-react";
import type { TimerState } from "@/lib/timer/ws-client";
import { Button } from "@/components/ui/button";

interface TimerControlsProps {
  state: TimerState;
  isAdmin: boolean;
  onStart: () => void;
  onPause: () => void;
  onReset: () => void;
  onSkip: () => void;
  onEndSession: () => void;
}

export function TimerControls({
  state,
  isAdmin,
  onStart,
  onPause,
  onReset,
  onSkip,
  onEndSession,
}: TimerControlsProps) {
  if (!isAdmin) {
    return (
      <div className="text-muted-foreground text-center text-sm">
        Only the room admin can control the timer
      </div>
    );
  }

  return (
    <div className="flex flex-wrap items-center justify-center gap-3">
      {state.isRunning ? (
        <Button
          onClick={onPause}
          size="lg"
          variant="secondary"
          className="min-w-[120px]"
        >
          <Pause className="mr-2 h-5 w-5" />
          Pause
        </Button>
      ) : (
        <Button onClick={onStart} size="lg" className="min-w-[120px]">
          <Play className="mr-2 h-5 w-5" />
          Start
        </Button>
      )}

      <Button onClick={onReset} size="lg" variant="outline">
        <RotateCcw className="mr-2 h-5 w-5" />
        Reset
      </Button>

      <Button onClick={onSkip} size="lg" variant="outline">
        <SkipForward className="mr-2 h-5 w-5" />
        Skip
      </Button>

      <Button onClick={onEndSession} size="lg" variant="destructive">
        <Square className="mr-2 h-5 w-5" />
        End
      </Button>
    </div>
  );
}
