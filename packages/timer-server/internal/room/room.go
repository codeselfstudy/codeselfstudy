package room

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/codeselfstudy/timer-server/internal/presence"
	"github.com/codeselfstudy/timer-server/internal/protocol"
	"github.com/codeselfstudy/timer-server/internal/timer"
)

// Client represents a connected WebSocket client.
type Client struct {
	ID      string
	Name    string
	IsAdmin bool
	Conn    *websocket.Conn
	Send    chan []byte
}

// Room manages the state and connections for a timer room.
type Room struct {
	ID        string
	CreatorID string

	mu       sync.RWMutex
	clients  map[string]*Client
	timer    *timer.Timer
	presence *presence.Manager

	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	done       chan struct{}

	// Heartbeat ticker for timer updates
	heartbeat *time.Ticker
}

// New creates a new room.
func New(id, creatorID string, settings protocol.TimerSettings) *Room {
	r := &Room{
		ID:         id,
		CreatorID:  creatorID,
		clients:    make(map[string]*Client),
		timer:      timer.New(settings),
		presence:   presence.New(),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
	}

	// Set up callbacks
	r.timer.SetCallbacks(r.onTimerStateChange, r.onTimerPhaseChange)
	r.presence.SetCallback(r.onPresenceChange)

	return r
}

// Run starts the room's main event loop.
func (r *Room) Run() {
	// Start heartbeat (1 second intervals)
	r.heartbeat = time.NewTicker(1 * time.Second)
	defer r.heartbeat.Stop()

	for {
		select {
		case <-r.done:
			return

		case client := <-r.register:
			r.mu.Lock()
			r.clients[client.ID] = client
			r.mu.Unlock()
			r.presence.Join(client.ID, client.Name, client.IsAdmin)

			// Send current state to new client
			r.sendStateToClient(client)

		case client := <-r.unregister:
			r.mu.Lock()
			if _, ok := r.clients[client.ID]; ok {
				delete(r.clients, client.ID)
				close(client.Send)
			}
			r.mu.Unlock()
			r.presence.Leave(client.ID)

		case msg := <-r.broadcast:
			r.mu.RLock()
			for _, client := range r.clients {
				select {
				case client.Send <- msg:
				default:
					// Client send buffer full, skip
				}
			}
			r.mu.RUnlock()

		case <-r.heartbeat.C:
			// Only broadcast heartbeat if timer is running
			state := r.timer.State()
			if state.IsRunning {
				r.broadcastTimerState(state)
			}
		}
	}
}

// Stop closes the room and disconnects all clients.
func (r *Room) Stop() {
	r.timer.Stop()

	select {
	case <-r.done:
	default:
		close(r.done)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	for _, client := range r.clients {
		close(client.Send)
	}
	r.clients = make(map[string]*Client)
}

// Register adds a client to the room.
func (r *Room) Register(client *Client) {
	r.register <- client
}

// Unregister removes a client from the room.
func (r *Room) Unregister(client *Client) {
	r.unregister <- client
}

// HandleMessage processes an incoming message from a client.
func (r *Room) HandleMessage(client *Client, data []byte) {
	var msg protocol.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		r.sendError(client, protocol.ErrCodeInvalidMessage, "Invalid message format")
		return
	}

	switch msg.Type {
	case protocol.MsgPing:
		r.sendPong(client)

	case protocol.MsgAdminStart:
		if !r.isAdmin(client) {
			r.sendError(client, protocol.ErrCodeUnauthorized, "Only admin can control the timer")
			return
		}
		r.timer.Start()

	case protocol.MsgAdminPause:
		if !r.isAdmin(client) {
			r.sendError(client, protocol.ErrCodeUnauthorized, "Only admin can control the timer")
			return
		}
		r.timer.Pause()

	case protocol.MsgAdminReset:
		if !r.isAdmin(client) {
			r.sendError(client, protocol.ErrCodeUnauthorized, "Only admin can control the timer")
			return
		}
		r.timer.Reset()

	case protocol.MsgAdminSkip:
		if !r.isAdmin(client) {
			r.sendError(client, protocol.ErrCodeUnauthorized, "Only admin can control the timer")
			return
		}
		r.timer.Skip()

	case protocol.MsgAdminUpdateSettings:
		if !r.isAdmin(client) {
			r.sendError(client, protocol.ErrCodeUnauthorized, "Only admin can control the timer")
			return
		}
		var settings protocol.TimerSettings
		if payload, ok := msg.Payload["settings"]; ok {
			if data, err := json.Marshal(payload); err == nil {
				json.Unmarshal(data, &settings)
			}
		}
		r.timer.UpdateSettings(settings)

	case protocol.MsgAdminEndSession:
		if !r.isAdmin(client) {
			r.sendError(client, protocol.ErrCodeUnauthorized, "Only admin can control the timer")
			return
		}
		r.broadcastSessionEnded("Admin ended the session")
		r.Stop()
	}
}

// ClientCount returns the number of connected clients.
func (r *Room) ClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

// TimerState returns the current timer state.
func (r *Room) TimerState() protocol.TimerState {
	return r.timer.State()
}

func (r *Room) isAdmin(client *Client) bool {
	return client.IsAdmin || client.ID == r.CreatorID
}

func (r *Room) onTimerStateChange(state protocol.TimerState) {
	r.broadcastTimerState(state)
}

func (r *Room) onTimerPhaseChange(oldPhase, newPhase protocol.TimerPhase) {
	// Phase change is already broadcast via state change
	log.Printf("Room %s: phase changed from %s to %s", r.ID, oldPhase, newPhase)
}

func (r *Room) onPresenceChange(users []protocol.User, count int) {
	msg := protocol.NewMessage(protocol.MsgPresenceUpdate, map[string]any{
		"users": users,
		"count": count,
	})
	r.broadcastMessage(msg)
}

func (r *Room) broadcastTimerState(state protocol.TimerState) {
	msg := protocol.NewMessage(protocol.MsgTimerUpdate, map[string]any{
		"phase":       state.Phase,
		"remainingMs": state.RemainingMs,
		"isRunning":   state.IsRunning,
		"cycleNumber": state.CycleNumber,
		"settings":    state.Settings,
	})
	r.broadcastMessage(msg)
}

func (r *Room) broadcastSessionEnded(reason string) {
	msg := protocol.NewMessage(protocol.MsgSessionEnded, map[string]any{
		"reason": reason,
	})
	r.broadcastMessage(msg)
}

func (r *Room) broadcastMessage(msg protocol.Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	select {
	case r.broadcast <- data:
	default:
		log.Printf("Broadcast channel full for room %s", r.ID)
	}
}

func (r *Room) sendStateToClient(client *Client) {
	// Send timer state
	state := r.timer.State()
	timerMsg := protocol.NewMessage(protocol.MsgTimerUpdate, map[string]any{
		"phase":       state.Phase,
		"remainingMs": state.RemainingMs,
		"isRunning":   state.IsRunning,
		"cycleNumber": state.CycleNumber,
		"settings":    state.Settings,
	})
	r.sendToClient(client, timerMsg)

	// Send presence
	users := r.presence.Users()
	presenceMsg := protocol.NewMessage(protocol.MsgPresenceUpdate, map[string]any{
		"users": users,
		"count": len(users),
	})
	r.sendToClient(client, presenceMsg)
}

func (r *Room) sendToClient(client *Client, msg protocol.Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case client.Send <- data:
	default:
	}
}

func (r *Room) sendPong(client *Client) {
	msg := protocol.NewMessage(protocol.MsgPong, nil)
	r.sendToClient(client, msg)
}

func (r *Room) sendError(client *Client, code, message string) {
	msg := protocol.NewMessage(protocol.MsgError, map[string]any{
		"code":    code,
		"message": message,
	})
	r.sendToClient(client, msg)
}

// WriteLoop handles writing messages to the client's WebSocket.
func WriteLoop(ctx context.Context, client *Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-client.Send:
			if !ok {
				return
			}
			err := client.Conn.Write(ctx, websocket.MessageText, msg)
			if err != nil {
				return
			}
		}
	}
}
