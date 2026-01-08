package timer

import (
	"sync"
	"time"

	"github.com/codeselfstudy/timer-server/internal/protocol"
)

// Timer manages the countdown state for a room.
type Timer struct {
	mu sync.RWMutex

	settings    protocol.TimerSettings
	phase       protocol.TimerPhase
	remainingMs int64
	isRunning   bool
	cycleNumber int // 1-4, resets after long break

	lastTick time.Time
	ticker   *time.Ticker
	done     chan struct{}

	// Callback when state changes
	onStateChange func(protocol.TimerState)
	// Callback when phase transitions
	onPhaseChange func(protocol.TimerPhase, protocol.TimerPhase)
}

// New creates a new timer with the given settings.
func New(settings protocol.TimerSettings) *Timer {
	return &Timer{
		settings:    settings,
		phase:       protocol.PhaseFocus,
		remainingMs: settings.FocusDuration,
		cycleNumber: 1,
		done:        make(chan struct{}),
	}
}

// SetCallbacks sets the callback functions for state and phase changes.
func (t *Timer) SetCallbacks(onStateChange func(protocol.TimerState), onPhaseChange func(protocol.TimerPhase, protocol.TimerPhase)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onStateChange = onStateChange
	t.onPhaseChange = onPhaseChange
}

// State returns the current timer state.
func (t *Timer) State() protocol.TimerState {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return protocol.TimerState{
		Phase:       t.phase,
		RemainingMs: t.remainingMs,
		IsRunning:   t.isRunning,
		CycleNumber: t.cycleNumber,
		Settings:    t.settings,
	}
}

// Start begins or resumes the countdown.
func (t *Timer) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isRunning {
		return
	}

	t.isRunning = true
	t.lastTick = time.Now()

	// Start the ticker
	t.ticker = time.NewTicker(100 * time.Millisecond)
	go t.runTicker()

	t.notifyStateChange()
}

// Pause stops the countdown without resetting.
func (t *Timer) Pause() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isRunning {
		return
	}

	t.isRunning = false
	if t.ticker != nil {
		t.ticker.Stop()
		t.ticker = nil
	}

	t.notifyStateChange()
}

// Reset returns the timer to the start of the current phase.
func (t *Timer) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	wasRunning := t.isRunning
	if t.ticker != nil {
		t.ticker.Stop()
		t.ticker = nil
	}
	t.isRunning = false

	// Reset to start of current phase
	switch t.phase {
	case protocol.PhaseFocus:
		t.remainingMs = t.settings.FocusDuration
	case protocol.PhaseBreak:
		t.remainingMs = t.settings.BreakDuration
	case protocol.PhaseLongBreak:
		t.remainingMs = t.settings.LongBreakDuration
	case protocol.PhaseOvertime:
		// Reset from overtime goes back to focus
		t.phase = protocol.PhaseFocus
		t.remainingMs = t.settings.FocusDuration
	}

	t.notifyStateChange()

	// Restart if it was running
	if wasRunning {
		t.mu.Unlock()
		t.Start()
		t.mu.Lock()
	}
}

// Skip advances to the next phase.
func (t *Timer) Skip() {
	t.mu.Lock()
	defer t.mu.Unlock()

	oldPhase := t.phase
	t.advancePhase()
	t.notifyPhaseChange(oldPhase, t.phase)
	t.notifyStateChange()
}

// UpdateSettings changes the timer durations.
func (t *Timer) UpdateSettings(settings protocol.TimerSettings) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.settings = settings

	// If not running, update remaining time to match new settings
	if !t.isRunning {
		switch t.phase {
		case protocol.PhaseFocus:
			t.remainingMs = settings.FocusDuration
		case protocol.PhaseBreak:
			t.remainingMs = settings.BreakDuration
		case protocol.PhaseLongBreak:
			t.remainingMs = settings.LongBreakDuration
		}
	}

	t.notifyStateChange()
}

// Stop completely stops the timer and cleans up resources.
func (t *Timer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.isRunning = false
	if t.ticker != nil {
		t.ticker.Stop()
		t.ticker = nil
	}

	select {
	case <-t.done:
	default:
		close(t.done)
	}
}

func (t *Timer) runTicker() {
	for {
		select {
		case <-t.done:
			return
		case <-t.ticker.C:
			t.tick()
		}
	}
}

func (t *Timer) tick() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isRunning {
		return
	}

	now := time.Now()
	elapsed := now.Sub(t.lastTick).Milliseconds()
	t.lastTick = now

	if t.phase == protocol.PhaseOvertime {
		// In overtime, count up (negative remaining means overtime)
		t.remainingMs -= elapsed
	} else {
		t.remainingMs -= elapsed

		if t.remainingMs <= 0 {
			oldPhase := t.phase
			t.advancePhase()
			t.notifyPhaseChange(oldPhase, t.phase)
		}
	}
}

func (t *Timer) advancePhase() {
	switch t.phase {
	case protocol.PhaseFocus:
		// After focus, take a break
		if t.cycleNumber >= 4 {
			t.phase = protocol.PhaseLongBreak
			t.remainingMs = t.settings.LongBreakDuration
			t.cycleNumber = 0 // Will be incremented on next focus
		} else {
			t.phase = protocol.PhaseBreak
			t.remainingMs = t.settings.BreakDuration
		}
	case protocol.PhaseBreak, protocol.PhaseLongBreak:
		// After break, back to focus
		t.phase = protocol.PhaseFocus
		t.remainingMs = t.settings.FocusDuration
		t.cycleNumber++
	case protocol.PhaseOvertime:
		// From overtime, go to focus
		t.phase = protocol.PhaseFocus
		t.remainingMs = t.settings.FocusDuration
		t.cycleNumber = 1
	}
}

func (t *Timer) notifyStateChange() {
	if t.onStateChange != nil {
		state := protocol.TimerState{
			Phase:       t.phase,
			RemainingMs: t.remainingMs,
			IsRunning:   t.isRunning,
			CycleNumber: t.cycleNumber,
			Settings:    t.settings,
		}
		go t.onStateChange(state)
	}
}

func (t *Timer) notifyPhaseChange(oldPhase, newPhase protocol.TimerPhase) {
	if t.onPhaseChange != nil {
		go t.onPhaseChange(oldPhase, newPhase)
	}
}
