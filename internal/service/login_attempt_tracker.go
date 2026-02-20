package service

import (
	"sync"
	"time"
)

type LoginAttemptTracker interface {
	IsBlocked(key string, now time.Time) (bool, time.Duration)
	RegisterFailure(key string, now time.Time)
	Reset(key string)
}

type loginAttemptState struct {
	failures     []time.Time
	blockedUntil time.Time
}

type MemoryLoginAttemptTracker struct {
	mu          sync.Mutex
	maxAttempts int
	window      time.Duration
	lockout     time.Duration
	store       map[string]loginAttemptState
}

func NewMemoryLoginAttemptTracker(maxAttempts int, window, lockout time.Duration) *MemoryLoginAttemptTracker {
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	if window <= 0 {
		window = 10 * time.Minute
	}
	if lockout <= 0 {
		lockout = 15 * time.Minute
	}
	return &MemoryLoginAttemptTracker{
		maxAttempts: maxAttempts,
		window:      window,
		lockout:     lockout,
		store:       make(map[string]loginAttemptState),
	}
}

func (m *MemoryLoginAttemptTracker) IsBlocked(key string, now time.Time) (bool, time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.store[key]
	if !ok {
		return false, 0
	}
	if now.Before(state.blockedUntil) {
		return true, state.blockedUntil.Sub(now)
	}
	if !state.blockedUntil.IsZero() && !now.Before(state.blockedUntil) {
		state.blockedUntil = time.Time{}
		state.failures = nil
		m.store[key] = state
	}
	return false, 0
}

func (m *MemoryLoginAttemptTracker) RegisterFailure(key string, now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := m.store[key]
	if now.Before(state.blockedUntil) {
		m.store[key] = state
		return
	}

	cutoff := now.Add(-m.window)
	kept := state.failures[:0]
	for _, ts := range state.failures {
		if !ts.Before(cutoff) {
			kept = append(kept, ts)
		}
	}
	state.failures = append(kept, now)
	if len(state.failures) >= m.maxAttempts {
		state.blockedUntil = now.Add(m.lockout)
		state.failures = nil
	}
	m.store[key] = state
}

func (m *MemoryLoginAttemptTracker) Reset(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.store, key)
}
