package streaming

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// SessionState represents the state of a streaming session
type SessionState string

const (
	SessionStateActive   SessionState = "active"
	SessionStateComplete SessionState = "complete"
	SessionStateCanceled SessionState = "canceled"
	SessionStateError    SessionState = "error"
)

// Session represents an active streaming session
type Session struct {
	mu sync.RWMutex

	ID           string       `json:"id"`
	RequestID    string       `json:"request_id"`
	ProjectID    uuid.UUID    `json:"project_id"`
	State        SessionState `json:"state"`
	CreatedAt    time.Time    `json:"created_at"`
	LastActivity time.Time    `json:"last_activity"`
	MessageCount int          `json:"message_count"`
	Error        string       `json:"error,omitempty"`

	subscribers map[string]chan Event
	eventBuffer []Event
	bufferSize  int
}

// NewSession creates a new streaming session
func NewSession(requestID string, projectID uuid.UUID) *Session {
	now := time.Now()
	return &Session{
		ID:           uuid.New().String(),
		RequestID:    requestID,
		ProjectID:    projectID,
		State:        SessionStateActive,
		CreatedAt:    now,
		LastActivity: now,
		subscribers:  make(map[string]chan Event),
		eventBuffer:  make([]Event, 0),
		bufferSize:   100,
	}
}

// AddSubscriber adds a subscriber to the session
func (s *Session) AddSubscriber(subscriberID string) <-chan Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan Event, 50)
	s.subscribers[subscriberID] = ch
	s.LastActivity = time.Now()
	return ch
}

// RemoveSubscriber removes a subscriber from the session
func (s *Session) RemoveSubscriber(subscriberID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.subscribers[subscriberID]; ok {
		close(ch)
		delete(s.subscribers, subscriberID)
	}
	s.LastActivity = time.Now()
}

// HasSubscribers returns true if the session has active subscribers
func (s *Session) HasSubscribers() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.subscribers) > 0
}

// SubscriberCount returns the number of active subscribers
func (s *Session) SubscriberCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.subscribers)
}

// Emit sends an event to all subscribers
func (s *Session) Emit(event Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.LastActivity = time.Now()
	s.MessageCount++

	// Buffer the event
	if len(s.eventBuffer) >= s.bufferSize {
		s.eventBuffer = s.eventBuffer[1:]
	}
	s.eventBuffer = append(s.eventBuffer, event)

	// Send to all subscribers
	for _, ch := range s.subscribers {
		select {
		case ch <- event:
		default:
			// Channel full, skip
		}
	}
}

// Complete marks the session as complete
func (s *Session) Complete() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.State = SessionStateComplete
	s.LastActivity = time.Now()

	// Close all subscriber channels
	for id, ch := range s.subscribers {
		close(ch)
		delete(s.subscribers, id)
	}
}

// Cancel marks the session as canceled
func (s *Session) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.State = SessionStateCanceled
	s.LastActivity = time.Now()

	// Close all subscriber channels
	for id, ch := range s.subscribers {
		close(ch)
		delete(s.subscribers, id)
	}
}

// SetError marks the session as errored
func (s *Session) SetError(err string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.State = SessionStateError
	s.Error = err
	s.LastActivity = time.Now()
}

// IsActive returns true if the session is active
func (s *Session) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.State == SessionStateActive
}

// IsExpired returns true if the session has expired
func (s *Session) IsExpired(timeout time.Duration) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Since(s.CreatedAt) > timeout
}

// IsInactive returns true if the session has been inactive
func (s *Session) IsInactive(inactiveTime time.Duration) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Since(s.LastActivity) > inactiveTime
}

// GetStats returns session statistics
func (s *Session) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"id":               s.ID,
		"request_id":       s.RequestID,
		"project_id":       s.ProjectID.String(),
		"state":            s.State,
		"created_at":       s.CreatedAt,
		"last_activity":    s.LastActivity,
		"message_count":    s.MessageCount,
		"subscriber_count": len(s.subscribers),
		"error":            s.Error,
	}
}
