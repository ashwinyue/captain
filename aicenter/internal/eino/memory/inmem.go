package memory

import (
	"context"
	"sync"

	"github.com/cloudwego/eino/schema"
)

// InMemoryStore keeps messages in a process-local map.
// Suitable for development/testing; not shared across processes.
type InMemoryStore struct {
	mu   sync.RWMutex
	data map[string][]*schema.Message
}

// NewInMemoryStore creates a new in-memory store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string][]*schema.Message),
	}
}

// Write stores messages for the given session
func (s *InMemoryStore) Write(ctx context.Context, sessionID string, msgs []*schema.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[sessionID] = msgs
	return nil
}

// Read returns messages for the given session
func (s *InMemoryStore) Read(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	msgs, ok := s.data[sessionID]
	if !ok {
		return nil, nil
	}
	// Return a copy to prevent external modification
	result := make([]*schema.Message, len(msgs))
	copy(result, msgs)
	return result, nil
}

// Append adds messages to a session
func (s *InMemoryStore) Append(ctx context.Context, sessionID string, msgs ...*schema.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[sessionID] = append(s.data[sessionID], msgs...)
	return nil
}

// Delete removes a session's messages
func (s *InMemoryStore) Delete(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, sessionID)
	return nil
}

// GetWindowedMessages returns the last N messages for a session
func (s *InMemoryStore) GetWindowedMessages(ctx context.Context, sessionID string, windowSize int) ([]*schema.Message, error) {
	msgs, err := s.Read(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if len(msgs) <= windowSize {
		return msgs, nil
	}
	return msgs[len(msgs)-windowSize:], nil
}
