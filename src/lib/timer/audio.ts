import type { TimerPhase } from "@/lib/timer/ws-client";

const SOUND_PATHS = {
  focus: "/sounds/timer/focus.mp3",
  break: "/sounds/timer/break.mp3",
  longBreak: "/sounds/timer/long-break.mp3",
  overtime: "/sounds/timer/alarm.mp3", // Kept just in case, though unused
};

export class AudioController {
  private static instance: AudioController;
  private isMuted = false;
  private audioElements: Record<string, HTMLAudioElement> = {};

  private constructor() {
    // Load mute preference
    if (typeof window !== "undefined") {
      const saved = localStorage.getItem("timer_muted");
      this.isMuted = saved === "true";

      // Preload sounds with fallback
      Object.entries(SOUND_PATHS).forEach(([key, path]) => {
        const audio = new Audio(path);
        // If file fails to load, mark it as missing so we valid fallback
        audio.onerror = () => {
          this.audioElements[key] = null as any; // Will use fallback
        };
        this.audioElements[key] = audio;
      });
    }
  }

  public static getInstance(): AudioController {
    if (!AudioController.instance) {
      AudioController.instance = new AudioController();
    }
    return AudioController.instance;
  }

  /**
   * Plays a fallback beep/tone using Web Audio API if file is missing
   */
  private playFallback(phase: TimerPhase) {
    if (typeof window === "undefined") return;

    try {
      const AudioContext =
        window.AudioContext || (window as any).webkitAudioContext;
      if (!AudioContext) return;

      const ctx = new AudioContext();
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      gain.connect(ctx.destination);

      const now = ctx.currentTime;

      // Simple sound design based on phase
      switch (phase) {
        case "focus":
          // Energetic ascending chime
          osc.type = "sine";
          osc.frequency.setValueAtTime(440, now);
          osc.frequency.exponentialRampToValueAtTime(880, now + 0.1);
          gain.gain.setValueAtTime(0.1, now);
          gain.gain.exponentialRampToValueAtTime(0.01, now + 0.5);
          osc.start(now);
          osc.stop(now + 0.5);
          break;

        case "break":
        case "longBreak":
          // Relaxing descending tone
          osc.type = "sine";
          osc.frequency.setValueAtTime(440, now);
          osc.frequency.exponentialRampToValueAtTime(220, now + 0.3);
          gain.gain.setValueAtTime(0.1, now);
          gain.gain.exponentialRampToValueAtTime(0.01, now + 0.5);
          osc.start(now);
          osc.stop(now + 0.5);
          break;

        case "overtime":
          // Urgent double beep
          osc.type = "square";
          osc.frequency.setValueAtTime(880, now);
          gain.gain.setValueAtTime(0.1, now);
          gain.gain.setValueAtTime(0, now + 0.1);
          gain.gain.setValueAtTime(0.1, now + 0.15);
          gain.gain.exponentialRampToValueAtTime(0.01, now + 0.3);
          osc.start(now);
          osc.stop(now + 0.3);
          break;
      }
    } catch (e) {
      console.error("Web Audio fallback failed:", e);
    }
  }

  public play(phase: TimerPhase) {
    if (this.isMuted) return;

    const audio = this.audioElements[phase];
    if (audio && audio.error === null && audio.readyState >= 2) {
      // Audio likely loaded
      audio.currentTime = 0;
      audio.play().catch(() => this.playFallback(phase));
    } else {
      // File missing or not ready, use fallback
      this.playFallback(phase);
    }
  }

  public setMuted(muted: boolean) {
    this.isMuted = muted;
    localStorage.setItem("timer_muted", String(muted));
  }

  public getMuted(): boolean {
    return this.isMuted;
  }

  // Method to unlock audio context on first user interaction if needed
  // (Browser require interaction before playing audio)
  public static async initAudioContext() {
    // Helper for components to call on first click
    // React usually handles this well with event handlers
  }
}
