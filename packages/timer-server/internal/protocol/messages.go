package protocol

import "time"

// MessageType represents the type of WebSocket message.
type MessageType string

const (
	// Client -> Server
	MsgJoin               MessageType = "JOIN"
	MsgPing               MessageType = "PING"
	MsgAdminStart         MessageType = "ADMIN_START"
	MsgAdminPause         MessageType = "ADMIN_PAUSE"
	MsgAdminReset         MessageType = "ADMIN_RESET"
	MsgAdminSkip          MessageType = "ADMIN_SKIP"
	MsgAdminEndSession    MessageType = "ADMIN_END_SESSION"
	MsgAdminUpdateSettings MessageType = "ADMIN_UPDATE_SETTINGS"

	// Server -> Client
	MsgTimerUpdate    MessageType = "TIMER_UPDATE"
	MsgPresenceUpdate MessageType = "PRESENCE_UPDATE"
	MsgError          MessageType = "ERROR"
	MsgSessionEnded   MessageType = "SESSION_ENDED"
	MsgPong           MessageType = "PONG"
)

// Message is the base WebSocket message structure.
type Message struct {
	Type      MessageType    `json:"type"`
	Payload   map[string]any `json:"payload,omitempty"`
	Timestamp int64          `json:"timestamp"`
}

// NewMessage creates a new message with the current timestamp.
func NewMessage(msgType MessageType, payload map[string]any) Message {
	return Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}
}

// TimerPhase represents the current phase of the timer.
type TimerPhase string

const (
	PhaseFocus     TimerPhase = "focus"
	PhaseBreak     TimerPhase = "break"
	PhaseLongBreak TimerPhase = "longBreak"
	PhaseOvertime  TimerPhase = "overtime"
)

// TimerSettings holds the configurable durations for a timer.
type TimerSettings struct {
	FocusDuration     int64 `json:"focusDuration"`     // in milliseconds
	BreakDuration     int64 `json:"breakDuration"`     // in milliseconds
	LongBreakDuration int64 `json:"longBreakDuration"` // in milliseconds
}

// DefaultSettings returns the standard Pomodoro settings.
func DefaultSettings() TimerSettings {
	return TimerSettings{
		FocusDuration:     25 * 60 * 1000, // 25 minutes
		BreakDuration:     5 * 60 * 1000,  // 5 minutes
		LongBreakDuration: 15 * 60 * 1000, // 15 minutes
	}
}

// TimerState represents the current state of a timer.
type TimerState struct {
	Phase       TimerPhase    `json:"phase"`
	RemainingMs int64         `json:"remainingMs"`
	IsRunning   bool          `json:"isRunning"`
	CycleNumber int           `json:"cycleNumber"` // 1-4 before long break
	Settings    TimerSettings `json:"settings"`
}

// User represents a connected user in a room.
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsAdmin  bool   `json:"isAdmin"`
	JoinedAt int64  `json:"joinedAt"`
}

// JoinPayload is the payload for JOIN messages.
type JoinPayload struct {
	Token string `json:"token"` // JWT for admin, session ID for guests
	Name  string `json:"name"`  // Display name (for guests)
}

// ErrorPayload is the payload for ERROR messages.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error codes
const (
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeRoomNotFound   = "ROOM_NOT_FOUND"
	ErrCodeInvalidMessage = "INVALID_MESSAGE"
	ErrCodeInternalError  = "INTERNAL_ERROR"
)
