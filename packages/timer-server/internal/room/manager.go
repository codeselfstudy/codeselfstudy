package room

import (
	"sync"

	"github.com/codeselfstudy/timer-server/internal/protocol"
)

// Manager handles multiple rooms.
type Manager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

// NewManager creates a new room manager.
func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

// Get returns a room by ID.
func (m *Manager) Get(id string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.rooms[id]
	return r, ok
}

// Create creates a new room with the given settings.
func (m *Manager) Create(id, creatorID string, settings protocol.TimerSettings) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if room already exists
	if r, ok := m.rooms[id]; ok {
		return r
	}

	r := New(id, creatorID, settings)
	m.rooms[id] = r

	// Start the room's event loop
	go r.Run()

	return r
}

// Delete removes a room.
func (m *Manager) Delete(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if r, ok := m.rooms[id]; ok {
		r.Stop()
		delete(m.rooms, id)
	}
}

// Count returns the number of active rooms.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms)
}

// List returns all room IDs.
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.rooms))
	for id := range m.rooms {
		ids = append(ids, id)
	}
	return ids
}

// CleanupEmpty removes rooms with no clients.
func (m *Manager) CleanupEmpty() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleaned := 0
	for id, r := range m.rooms {
		if r.ClientCount() == 0 {
			r.Stop()
			delete(m.rooms, id)
			cleaned++
		}
	}
	return cleaned
}
