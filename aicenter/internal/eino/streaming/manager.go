package streaming

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ManagerConfig configures the stream manager
type ManagerConfig struct {
	// SessionTimeout is the maximum duration a session can exist
	SessionTimeout time.Duration
	// InactiveTimeout is the duration after which an inactive session is cleaned up
	InactiveTimeout time.Duration
	// CleanupInterval is the interval for running cleanup
	CleanupInterval time.Duration
}

// DefaultManagerConfig returns default configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		SessionTimeout:  time.Hour,
		InactiveTimeout: 10 * time.Minute,
		CleanupInterval: 5 * time.Minute,
	}
}

// Manager manages streaming sessions
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	config   *ManagerConfig

	// Maps request ID to session ID for lookup
	requestToSession map[string]string

	stopCleanup chan struct{}
	wg          sync.WaitGroup
}

// NewManager creates a new stream manager
func NewManager(config *ManagerConfig) *Manager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	return &Manager{
		sessions:         make(map[string]*Session),
		requestToSession: make(map[string]string),
		config:           config,
		stopCleanup:      make(chan struct{}),
	}
}

// Start starts the stream manager background tasks
func (m *Manager) Start() {
	m.wg.Add(1)
	go m.cleanupLoop()
}

// Stop stops the stream manager
func (m *Manager) Stop() {
	close(m.stopCleanup)
	m.wg.Wait()

	// Close all sessions
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, session := range m.sessions {
		session.Cancel()
	}
	m.sessions = make(map[string]*Session)
	m.requestToSession = make(map[string]string)
}

// CreateSession creates a new streaming session
func (m *Manager) CreateSession(requestID string, projectID uuid.UUID) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := NewSession(requestID, projectID)
	m.sessions[session.ID] = session
	m.requestToSession[requestID] = session.ID

	return session
}

// GetSession returns a session by ID
func (m *Manager) GetSession(sessionID string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[sessionID]
	return session, ok
}

// GetSessionByRequest returns a session by request ID
func (m *Manager) GetSessionByRequest(requestID string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessionID, ok := m.requestToSession[requestID]
	if !ok {
		return nil, false
	}

	session, ok := m.sessions[sessionID]
	return session, ok
}

// DeleteSession removes a session
func (m *Manager) DeleteSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return
	}

	session.Cancel()
	delete(m.sessions, sessionID)
	delete(m.requestToSession, session.RequestID)
}

// ListSessions returns all active session IDs
func (m *Manager) ListSessions() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	return ids
}

// SessionCount returns the number of active sessions
func (m *Manager) SessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// GetStats returns manager statistics
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	activeSessions := 0
	totalSubscribers := 0

	for _, session := range m.sessions {
		if session.IsActive() {
			activeSessions++
		}
		totalSubscribers += session.SubscriberCount()
	}

	return map[string]interface{}{
		"total_sessions":    len(m.sessions),
		"active_sessions":   activeSessions,
		"total_subscribers": totalSubscribers,
	}
}

func (m *Manager) cleanupLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCleanup:
			return
		case <-ticker.C:
			m.cleanup()
		}
	}
}

func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	toDelete := make([]string, 0)

	for id, session := range m.sessions {
		// Remove expired sessions
		if session.IsExpired(m.config.SessionTimeout) {
			toDelete = append(toDelete, id)
			continue
		}

		// Remove inactive completed/errored sessions
		if !session.IsActive() && session.IsInactive(m.config.InactiveTimeout) {
			toDelete = append(toDelete, id)
			continue
		}

		// Remove inactive sessions without subscribers
		if session.IsInactive(m.config.InactiveTimeout) && !session.HasSubscribers() {
			toDelete = append(toDelete, id)
		}
	}

	for _, id := range toDelete {
		session := m.sessions[id]
		session.Cancel()
		delete(m.sessions, id)
		delete(m.requestToSession, session.RequestID)
	}
}

// EmitToSession emits an event to a specific session
func (m *Manager) EmitToSession(sessionID string, event Event) bool {
	m.mu.RLock()
	session, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok || !session.IsActive() {
		return false
	}

	session.Emit(event)
	return true
}

// Subscribe subscribes to a session's events
func (m *Manager) Subscribe(ctx context.Context, sessionID string) (<-chan Event, string, error) {
	m.mu.RLock()
	session, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return nil, "", ErrSessionNotFound
	}

	subscriberID := uuid.New().String()
	ch := session.AddSubscriber(subscriberID)

	return ch, subscriberID, nil
}

// Unsubscribe unsubscribes from a session
func (m *Manager) Unsubscribe(sessionID, subscriberID string) {
	m.mu.RLock()
	session, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if ok {
		session.RemoveSubscriber(subscriberID)
	}
}
