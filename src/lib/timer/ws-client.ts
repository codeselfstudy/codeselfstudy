import { useCallback, useEffect, useRef, useState } from "react";

// Message types matching the Go server
export type MessageType =
  | "JOIN"
  | "PING"
  | "PONG"
  | "ADMIN_START"
  | "ADMIN_PAUSE"
  | "ADMIN_RESET"
  | "ADMIN_SKIP"
  | "ADMIN_END_SESSION"
  | "ADMIN_UPDATE_SETTINGS"
  | "TIMER_UPDATE"
  | "PRESENCE_UPDATE"
  | "ERROR"
  | "SESSION_ENDED";

export interface WSMessage {
  type: MessageType;
  payload?: Record<string, unknown>;
  timestamp: number;
}

export type TimerPhase = "focus" | "break" | "longBreak" | "overtime";

export interface TimerSettings {
  focusDuration: number;
  breakDuration: number;
  longBreakDuration: number;
}

export interface TimerState {
  phase: TimerPhase;
  remainingMs: number;
  isRunning: boolean;
  cycleNumber: number;
  settings: TimerSettings;
}

export interface User {
  id: string;
  name: string;
  isAdmin: boolean;
  joinedAt: number;
}

export interface PresenceState {
  users: Array<User>;
  count: number;
}

export type ConnectionStatus =
  | "connecting"
  | "connected"
  | "disconnected"
  | "error";

interface UseTimerWSOptions {
  roomId: string;
  token: string;
  name?: string;
  onError?: (code: string, message: string) => void;
  onSessionEnded?: (reason: string) => void;
}

interface UseTimerWSReturn {
  status: ConnectionStatus;
  timerState: TimerState | null;
  presence: PresenceState | null;
  isAdmin: boolean;
  start: () => void;
  pause: () => void;
  reset: () => void;
  skip: () => void;
  endSession: () => void;
  updateSettings: (settings: Partial<TimerSettings>) => void;
}

const DEFAULT_TIMER_STATE: TimerState = {
  phase: "focus",
  remainingMs: 25 * 60 * 1000,
  isRunning: false,
  cycleNumber: 1,
  settings: {
    focusDuration: 25 * 60 * 1000,
    breakDuration: 5 * 60 * 1000,
    longBreakDuration: 15 * 60 * 1000,
  },
};

export function useTimerWS(options: UseTimerWSOptions): UseTimerWSReturn {
  const { roomId, token, name, onError, onSessionEnded } = options;

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(
    null
  );

  const [status, setStatus] = useState<ConnectionStatus>("connecting");
  const [timerState, setTimerState] = useState<TimerState | null>(null);
  const [presence, setPresence] = useState<PresenceState | null>(null);
  const [isAdmin, setIsAdmin] = useState(false);

  const send = useCallback(
    (type: MessageType, payload?: Record<string, unknown>) => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        const message: WSMessage = {
          type,
          payload,
          timestamp: Date.now(),
        };
        wsRef.current.send(JSON.stringify(message));
      }
    },
    []
  );

  const connect = useCallback(() => {
    // Determine WebSocket URL based on current location
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/ws/${roomId}`;

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      // Send JOIN message immediately after connection
      send("JOIN", { token, name });
      setStatus("connected");
    };

    ws.onmessage = (event) => {
      try {
        const message: WSMessage = JSON.parse(event.data);

        switch (message.type) {
          case "TIMER_UPDATE":
            setTimerState(message.payload as unknown as TimerState);
            break;

          case "PRESENCE_UPDATE": {
            const presenceData = message.payload as unknown as PresenceState;
            setPresence(presenceData);
            // Check if current user is admin based on token
            // The server validates this, we just use it for UI hints
            break;
          }

          case "ERROR":
            onError?.(
              message.payload?.code as string,
              message.payload?.message as string
            );
            break;

          case "SESSION_ENDED":
            onSessionEnded?.(message.payload?.reason as string);
            ws.close();
            break;

          case "PONG":
            // Heartbeat response, ignore
            break;
        }
      } catch (e) {
        console.error("Failed to parse WebSocket message:", e);
      }
    };

    ws.onclose = () => {
      setStatus("disconnected");
      // Attempt to reconnect after 3 seconds
      reconnectTimeoutRef.current = setTimeout(() => {
        if (wsRef.current === ws) {
          setStatus("connecting");
          connect();
        }
      }, 3000);
    };

    ws.onerror = () => {
      setStatus("error");
    };
  }, [roomId, token, name, send, onError, onSessionEnded]);

  // Connect on mount
  useEffect(() => {
    connect();

    return () => {
      if (reconnectTimeoutRef.current !== null) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current !== null) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [connect]);

  // Ping every 30 seconds to keep connection alive
  useEffect(() => {
    const pingInterval = setInterval(() => {
      send("PING");
    }, 30000);

    return () => clearInterval(pingInterval);
  }, [send]);

  // Admin check based on token (JWT contains isAdmin claim)
  useEffect(() => {
    try {
      // Decode JWT payload (without verification - server does that)
      const [, payloadBase64] = token.split(".");
      if (payloadBase64) {
        const payload = JSON.parse(atob(payloadBase64));
        setIsAdmin(payload.isAdmin === true);
      }
    } catch {
      setIsAdmin(false);
    }
  }, [token]);

  // Admin controls
  const start = useCallback(() => send("ADMIN_START"), [send]);
  const pause = useCallback(() => send("ADMIN_PAUSE"), [send]);
  const reset = useCallback(() => send("ADMIN_RESET"), [send]);
  const skip = useCallback(() => send("ADMIN_SKIP"), [send]);
  const endSession = useCallback(() => send("ADMIN_END_SESSION"), [send]);

  const updateSettings = useCallback(
    (settings: Partial<TimerSettings>) => {
      send("ADMIN_UPDATE_SETTINGS", { settings });
    },
    [send]
  );

  return {
    status,
    timerState: timerState ?? DEFAULT_TIMER_STATE,
    presence,
    isAdmin,
    start,
    pause,
    reset,
    skip,
    endSession,
    updateSettings,
  };
}
