package presence

import (
	"sync"
	"time"

	"github.com/codeselfstudy/timer-server/internal/protocol"
)

// Manager tracks connected users in a room.
type Manager struct {
	mu    sync.RWMutex
	users map[string]*protocol.User

	// Callback when presence changes
	onChange func([]protocol.User, int)
}

// New creates a new presence manager.
func New() *Manager {
	return &Manager{
		users: make(map[string]*protocol.User),
	}
}

// SetCallback sets the callback for presence changes.
func (m *Manager) SetCallback(onChange func([]protocol.User, int)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onChange = onChange
}

// Join adds a user to the room.
func (m *Manager) Join(id, name string, isAdmin bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.users[id] = &protocol.User{
		ID:       id,
		Name:     name,
		IsAdmin:  isAdmin,
		JoinedAt: time.Now().UnixMilli(),
	}

	m.notifyChange()
}

// Leave removes a user from the room.
func (m *Manager) Leave(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.users, id)
	m.notifyChange()
}

// Users returns all connected users.
func (m *Manager) Users() []protocol.User {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]protocol.User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, *u)
	}
	return users
}

// Count returns the number of connected users.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.users)
}

// Get returns a user by ID.
func (m *Manager) Get(id string) (*protocol.User, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return nil, false
	}
	return u, true
}

// IsAdmin checks if a user is an admin.
func (m *Manager) IsAdmin(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[id]
	return ok && u.IsAdmin
}

func (m *Manager) notifyChange() {
	if m.onChange == nil {
		return
	}

	users := make([]protocol.User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, *u)
	}

	go m.onChange(users, len(users))
}
