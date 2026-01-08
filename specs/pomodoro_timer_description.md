# Functional Specification: Group Pomodoro Timer

## 1. Overview

A real-time collaborative Pomodoro timer that allows a room admin to run synchronized focus/break sessions with participants. The timer is visible to all room members and controlled exclusively by the admin.

## 2. Architecture

### 2.1 System Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Fly.io Container (256MB)                 │
│                                                             │
│  ┌─────────────────────┐    ┌─────────────────────────┐    │
│  │   TanStack Start    │    │   Go WebSocket Server   │    │
│  │   (Bun/Nitro)       │    │   (coder/websocket)     │    │
│  │                     │    │                         │    │
│  │  - React Frontend   │    │  - Timer state (memory) │    │
│  │  - Room creation    │    │  - WS connections       │    │
│  │  - JWT signing      │    │  - Presence tracking    │    │
│  │  - DB operations    │    │  - Periodic DB sync     │    │
│  │                     │    │                         │    │
│  │  Port: 8080         │    │  Port: 8081 (internal)  │    │
│  └─────────────────────┘    └─────────────────────────┘    │
│           │                            │                    │
│           └──────────┬─────────────────┘                    │
│                      │                                      │
│              ┌───────▼───────┐                              │
│              │    Turso      │                              │
│              │   (SQLite)    │                              │
│              └───────────────┘                              │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Request Routing

All external traffic hits TanStack Start on port 8080. WebSocket connections are proxied internally:

```
Client → Fly.io (443) → TanStack Start (8080)
                              │
                              ├── /ws/* → proxy → Go Server (8081)
                              └── /*    → handle directly
```

TanStack Start uses Nitro's proxy capabilities to forward `/ws/*` to the Go server running on localhost:8081. This keeps deployment simple (single exposed port) and allows the Node app to handle authentication before upgrading to WebSocket.

### 2.3 Code Location

```
packages/
└── timer-server/           # Go WebSocket server
    ├── cmd/
    │   └── server/
    │       └── main.go
    ├── internal/
    │   ├── room/           # Room management
    │   ├── timer/          # Timer logic
    │   ├── presence/       # User presence
    │   └── protocol/       # WS message types
    ├── go.mod
    └── go.sum

src/
├── routes/
│   └── timer/
│       ├── index.tsx       # Timer landing/create page
│       └── $roomId.tsx     # Room view
├── components/
│   └── timer/
│       ├── timer-display.tsx
│       ├── timer-controls.tsx
│       ├── presence-list.tsx
│       └── room-settings.tsx
└── lib/
    └── timer/
        ├── jwt.ts          # JWT signing for admin auth
        └── ws-client.ts    # WebSocket client wrapper
```

## 3. Core User Flow

### 3.1 Room Creation (Authenticated Users Only)

1. User navigates to `/timer`
2. User clicks "Create Room"
3. System generates short room code (e.g., `abc123`)
4. Optional: User sets custom slug (e.g., `team-standup`)
5. User configures timer defaults:
   - Focus duration (default: 25m)
   - Short break (default: 5m)
   - Long break (default: 15m)
6. Room created, user redirected to `/timer/{roomId}`
7. TanStack Start issues JWT containing: `{userId, roomId, isAdmin: true}`

### 3.2 Joining a Room

**Authenticated users:**

1. Navigate to `/timer/{roomId}`
2. Connect to WebSocket with WorkOS session
3. Display name from WorkOS profile

**Anonymous users:**

1. Navigate to `/timer/{roomId}`
2. System generates session cookie if not present
3. Auto-generate name: `adjective_noun_number` (e.g., `stormy_salad_4329`)
4. Connect to WebSocket with session cookie

### 3.3 Timer Operation

1. **Setup**: Admin sees timer at initial duration with controls visible
2. **Start**: Admin clicks Start; all participants see synchronized countdown
3. **Countdown**: Timer counts down, 1-second heartbeat syncs all clients
4. **Transition** (at 00:00):
   - Audible chime plays (different sounds for focus vs. break end)
   - Visual theme changes (green/neutral → red for overtime)
   - Auto-advance to next phase (focus → break → focus...)
5. **Controls**: Admin can pause, reset, or skip to next phase at any time

## 4. Functional Requirements

### 4.1 Timer Controls (Admin Only)

| Control     | Behavior                              |
| ----------- | ------------------------------------- |
| Start       | Begin countdown from current duration |
| Pause       | Freeze timer at current time          |
| Reset       | Return to start of current phase      |
| Skip        | Jump to next phase (focus ↔ break)    |
| End Session | Close the room for all participants   |

### 4.2 Timer Display

- **Format**: `MM:SS` (under 1 hour) or `HH:MM:SS`
- **Phases**:
  - Focus: Neutral/green theme
  - Break: Blue/calm theme
  - Overtime: Red/alert theme (after 00:00)
- **Cycle indicator**: "Focus 2/4" showing progress toward long break

### 4.3 Audio

- Focus start: Energetic chime
- Break start: Relaxing tone
- Overtime: Urgent alert
- Mute toggle per user (persisted in localStorage)

### 4.4 Presence

- Display list of all connected users
- Show admin badge next to room owner
- Update presence in real-time (join/leave)
- Show "N participants" summary when list is long

## 5. Room URLs

Two URL formats supported:

1. **Auto-generated**: `/timer/abc123` (6-char alphanumeric)
2. **Custom slug**: `/timer/team-standup` (user-chosen, uniqueness enforced)

Slugs must be:

- 3-50 characters
- Alphanumeric + hyphens only
- Not conflict with reserved words or existing codes

## 6. WebSocket Protocol

### 6.1 Message Format

```typescript
interface WSMessage {
  type: MessageType;
  payload: Record<string, unknown>;
  timestamp: number; // Unix ms
}
```

### 6.2 Message Types

**Client → Server:**

```typescript
// Admin commands
{ type: "ADMIN_START" }
{ type: "ADMIN_PAUSE" }
{ type: "ADMIN_RESET" }
{ type: "ADMIN_SKIP" }
{ type: "ADMIN_END_SESSION" }
{ type: "ADMIN_UPDATE_SETTINGS", payload: { focusDuration, breakDuration, longBreakDuration } }

// All users
{ type: "JOIN", payload: { token: string } }  // JWT for admin, session for guests
{ type: "PING" }
```

**Server → Client:**

```typescript
// Timer state (broadcast every 1s + on state change)
{
  type: "TIMER_UPDATE",
  payload: {
    phase: "focus" | "break" | "longBreak" | "overtime",
    remainingMs: number,
    isRunning: boolean,
    cycleNumber: number,      // 1-4 before long break
    settings: { focusDuration, breakDuration, longBreakDuration }
  }
}

// Presence updates
{
  type: "PRESENCE_UPDATE",
  payload: {
    users: [{ id, name, isAdmin, joinedAt }],
    count: number
  }
}

// Errors
{ type: "ERROR", payload: { code: string, message: string } }

// Session events
{ type: "SESSION_ENDED", payload: { reason: string } }
```

### 6.3 Sync Strategy

- **Heartbeat**: Server broadcasts `TIMER_UPDATE` every 1 second while timer is running
- **State changes**: Immediate broadcast on start/pause/reset/skip/phase-change
- **Reconnection**: Client requests full state on reconnect; server sends current `TIMER_UPDATE` + `PRESENCE_UPDATE`

## 7. Authentication & Authorization

### 7.1 Room Creation

- Requires WorkOS authentication
- Creates DB record with `creatorId`

### 7.2 Admin Verification

- TanStack Start signs JWT on room creation/rejoin:
  ```typescript
  {
    sub: userId,
    roomId: string,
    isAdmin: boolean,
    exp: number  // Short-lived, e.g., 1 hour
  }
  ```
- Go server verifies JWT signature (shared secret or public key)
- Only requests with valid admin JWT can execute `ADMIN_*` commands

### 7.3 Guest Access

- Session cookie generated on first visit (UUID)
- Name generated from cookie ID: `adjective_noun_hash`
- Guests can view timer and presence, cannot control

## 8. Data Persistence

### 8.1 Database Schema (Drizzle/SQLite)

```typescript
// Rooms table
// NOTE: The actual implementation uses snake_case for column names
export const timerRooms = sqliteTable("timer_rooms", {
  id: text("id").primaryKey(), // Short code or slug
  creatorId: text("creator_id").notNull(), // WorkOS user ID
  slug: text("slug").unique(), // Optional custom slug
  settings: text("settings", { mode: "json" }), // Timer durations
  createdAt: integer("created_at", { mode: "timestamp" }),
  lastActiveAt: integer("last_active_at", { mode: "timestamp" }),
});

// Room state (for recovery)
export const timerRoomState = sqliteTable("timer_room_state", {
  roomId: text("room_id")
    .primaryKey()
    .references(() => timerRooms.id),
  phase: text("phase").notNull(), // focus, break, longBreak
  remainingMs: integer("remaining_ms").notNull(),
  isRunning: integer("is_running", { mode: "boolean" }),
  cycleNumber: integer("cycle_number"),
  updatedAt: integer("updated_at", { mode: "timestamp" }),
});
```

### 8.2 State Sync

- **In-memory**: Go server holds authoritative timer state for active rooms
- **Periodic sync**: Write to `timer_room_state` every 60 seconds (or on significant events)
- **Recovery**: On Go server restart, load active rooms from DB

## 9. UI Design

### 9.1 Layout

```
┌─────────────────────────────────────────┐
│  [Logo]     Room: team-standup    [Mute]│
├─────────────────────────────────────────┤
│                                         │
│              ┌─────────┐                │
│              │  24:37  │                │
│              └─────────┘                │
│            Focus (2/4)                  │
│                                         │
│     [Start]  [Reset]  [Skip]            │  ← Admin only
│                                         │
├─────────────────────────────────────────┤
│  Participants (5)                       │
│  👑 Alice (you)                         │
│  stormy_salad_4329                      │
│  quiet_mountain_8821                    │
│  Bob                                    │
│  Carol                                  │
└─────────────────────────────────────────┘
```

### 9.2 Responsive Behavior

- **Desktop**: Full layout with sidebar presence list
- **Mobile**: Timer dominates, presence collapsed to count badge
- **Projected/Kiosk**: Timer-only mode, minimal chrome

## 10. Deployment

### 10.1 Dockerfile Changes

```dockerfile
# --- Go build stage ---
FROM golang:1.23-alpine AS go-builder
WORKDIR /app/packages/timer-server
COPY packages/timer-server/go.mod packages/timer-server/go.sum ./
RUN go mod download
COPY packages/timer-server/ ./
RUN CGO_ENABLED=0 go build -o /timer-server ./cmd/server

# --- Final stage (add to existing) ---
COPY --from=go-builder /timer-server /app/timer-server
```

### 10.2 Process Management

- Use a simple process supervisor (e.g., `s6-overlay` or shell script)
- Start both Bun app and Go server
- Health checks on both processes

### 10.3 fly.toml Updates

No changes needed to fly.toml services - TanStack Start handles all external traffic on port 8080 and proxies WebSocket connections internally. The Go server only listens on localhost:8081 and is not exposed externally.

## 11. Implementation Phases

### Phase 1: Foundation

- [ ] Set up Go project structure in `packages/timer-server/`
- [ ] Implement basic WebSocket server with room management
- [ ] Create database schema and migrations
- [ ] Build room creation flow in TanStack Start

### Phase 2: Timer Core

- [ ] Implement timer logic (countdown, phases, cycles)
- [ ] Build React timer display component
- [ ] Implement WebSocket client in React
- [ ] Add JWT signing/verification for admin auth

### Phase 3: Real-time Features

- [ ] Implement presence tracking
- [ ] Add heartbeat synchronization
- [ ] Build presence list UI
- [ ] Add audio notifications

### Phase 4: Polish & Deploy

- [ ] Custom slug support
- [ ] State recovery from DB
- [ ] Responsive design refinements
- [ ] Update Dockerfile and fly.toml
- [ ] Testing and documentation

## 12. Future Enhancements (Post-MVP)

- Multiple admins per room
- Timer history/statistics
- Custom themes
- Slack/Discord integration
- Calendar integration for scheduled sessions
