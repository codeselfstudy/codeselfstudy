import { useEffect, useMemo, useState } from "react";
import { Volume2, VolumeX } from "lucide-react";
import type { TimerPhase, TimerState } from "@/lib/timer/ws-client";
import { cn } from "@/lib/utils";
import { AudioController } from "@/lib/timer/audio";

interface TimerDisplayProps {
  state: TimerState;
  className?: string;
}

function formatTime(ms: number): string {
  const totalSeconds = Math.max(0, Math.floor(ms / 1000));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
  }
  return `${minutes.toString().padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
}

function getPhaseLabel(phase: TimerPhase, cycleNumber: number): string {
  switch (phase) {
    case "focus":
      return `Focus (${cycleNumber}/4)`;
    case "break":
      return "Short Break";
    case "longBreak":
      return "Long Break";
    case "overtime":
      return "Overtime";
  }
}

function getPhaseStyles(phase: TimerPhase): string {
  switch (phase) {
    case "focus":
      // Subtle green for focus
      return "bg-green-50 text-green-900 border-green-100 dark:bg-green-950/30 dark:text-green-100 dark:border-green-900";
    case "break":
      // Subtle blue for break
      return "bg-blue-50 text-blue-900 border-blue-100 dark:bg-blue-950/30 dark:text-blue-100 dark:border-blue-900";
    case "longBreak":
      // Subtle purple for long break
      return "bg-indigo-50 text-indigo-900 border-indigo-100 dark:bg-indigo-950/30 dark:text-indigo-100 dark:border-indigo-900";
    case "overtime":
      // Red for overtime/alert
      return "bg-red-50 text-red-900 border-red-100 dark:bg-red-950/30 dark:text-red-100 dark:border-red-900";
  }
}

export function TimerDisplay({ state, className }: TimerDisplayProps) {
  const [isMuted, setIsMuted] = useState(
    AudioController.getInstance().getMuted()
  );

  // Handle audio playback on phase changes
  useEffect(() => {
    // Only play if the timer is running or just finished
    if (state.isRunning || state.remainingMs === 0) {
      AudioController.getInstance().play(state.phase);
    }
  }, [state.phase]);

  const toggleMute = () => {
    const newState = !isMuted;
    setIsMuted(newState);
    AudioController.getInstance().setMuted(newState);
  };

  const timeDisplay = useMemo(() => {
    if (state.phase === "overtime") {
      // For overtime, remaining is negative (counting up)
      return formatTime(Math.abs(state.remainingMs));
    }
    return formatTime(state.remainingMs);
  }, [state.remainingMs, state.phase]);

  const phaseLabel = useMemo(
    () => getPhaseLabel(state.phase, state.cycleNumber),
    [state.phase, state.cycleNumber]
  );

  const phaseStyles = useMemo(() => getPhaseStyles(state.phase), [state.phase]);

  return (
    <div
      className={cn(
        "relative flex flex-col items-center justify-center rounded-xl border p-8 transition-colors duration-500",
        phaseStyles,
        className
      )}
    >
      <button
        onClick={toggleMute}
        className="absolute top-4 right-4 rounded-full p-2 transition-colors hover:bg-black/5 dark:hover:bg-white/10"
        aria-label={isMuted ? "Unmute" : "Mute"}
      >
        {isMuted ? (
          <VolumeX className="h-5 w-5 opacity-50" />
        ) : (
          <Volume2 className="h-5 w-5 opacity-50" />
        )}
      </button>

      <div
        className="font-mono text-[15vw] leading-none font-bold tracking-tight tabular-nums md:text-[12vw] lg:text-[10vw]"
        role="timer"
        aria-live="polite"
        aria-label={`${timeDisplay} remaining in ${phaseLabel}`}
      >
        {timeDisplay}
      </div>
      <div className="mt-4 text-xl font-medium opacity-80 md:text-2xl">
        {phaseLabel}
      </div>
      {!state.isRunning && state.remainingMs > 0 && (
        <div className="mt-2 text-sm opacity-60">Paused</div>
      )}
    </div>
  );
}
