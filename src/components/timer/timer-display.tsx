import { useMemo } from "react";
import type { TimerPhase, TimerState } from "@/lib/timer/ws-client";
import { cn } from "@/lib/utils";

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
      return "bg-background text-foreground";
    case "break":
      return "bg-blue-50 text-blue-900 dark:bg-blue-950 dark:text-blue-100";
    case "longBreak":
      return "bg-indigo-50 text-indigo-900 dark:bg-indigo-950 dark:text-indigo-100";
    case "overtime":
      return "bg-red-600 text-white dark:bg-red-700";
  }
}

export function TimerDisplay({ state, className }: TimerDisplayProps) {
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
        "flex flex-col items-center justify-center rounded-lg p-8 transition-colors duration-500",
        phaseStyles,
        className
      )}
    >
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
