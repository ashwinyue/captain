package memory

import (
	"context"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ManagerConfig configures the memory manager
type ManagerConfig struct {
	// WindowSize is the maximum number of messages to include in context
	WindowSize int
	// EnablePersistence enables database persistence
	EnablePersistence bool
	// DB is the database connection (required if EnablePersistence is true)
	DB *gorm.DB
	// ProjectID for scoping messages (required if EnablePersistence is true)
	ProjectID uuid.UUID
	// Store is an optional pre-existing store to use (takes precedence)
	Store Store
	// RedisStore for hot cache (optional, used with HybridStore)
	RedisStore *RedisStore
	// Summarizer for conversation compression (optional)
	Summarizer *Summarizer
}

// Manager manages conversation memory for agents
type Manager struct {
	store      Store
	windowSize int
	summarizer *Summarizer
}

// NewManager creates a new memory manager
func NewManager(cfg *ManagerConfig) *Manager {
	windowSize := cfg.WindowSize
	if windowSize <= 0 {
		windowSize = 10 // Default window size
	}

	var store Store
	if cfg.Store != nil {
		// Use provided store (shared across requests)
		store = cfg.Store
	} else if cfg.EnablePersistence && cfg.DB != nil {
		pgStore := NewPostgresStore(cfg.DB, cfg.ProjectID)
		// Use HybridStore if Redis is available
		if cfg.RedisStore != nil {
			store = NewHybridStore(cfg.RedisStore, pgStore)
		} else {
			store = pgStore
		}
	} else {
		store = NewInMemoryStore()
	}

	return &Manager{
		store:      store,
		windowSize: windowSize,
		summarizer: cfg.Summarizer,
	}
}

// GetHistory returns the conversation history for a session
func (m *Manager) GetHistory(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	return m.store.Read(ctx, sessionID)
}

// GetWindowedHistory returns the last N messages for a session
func (m *Manager) GetWindowedHistory(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	msgs, err := m.store.Read(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if len(msgs) <= m.windowSize {
		return msgs, nil
	}
	return msgs[len(msgs)-m.windowSize:], nil
}

// AddMessage adds a message to the conversation history
func (m *Manager) AddMessage(ctx context.Context, sessionID string, msg *schema.Message) error {
	return m.store.Append(ctx, sessionID, msg)
}

// AddMessages adds multiple messages to the conversation history
func (m *Manager) AddMessages(ctx context.Context, sessionID string, msgs ...*schema.Message) error {
	return m.store.Append(ctx, sessionID, msgs...)
}

// AddUserMessage adds a user message to the history
func (m *Manager) AddUserMessage(ctx context.Context, sessionID string, content string) error {
	return m.AddMessage(ctx, sessionID, schema.UserMessage(content))
}

// AddAssistantMessage adds an assistant message to the history
func (m *Manager) AddAssistantMessage(ctx context.Context, sessionID string, content string) error {
	return m.AddMessage(ctx, sessionID, schema.AssistantMessage(content, nil))
}

// ClearHistory clears the conversation history for a session
func (m *Manager) ClearHistory(ctx context.Context, sessionID string) error {
	return m.store.Delete(ctx, sessionID)
}

// BuildContextMessages builds the messages to include in the LLM context
// This includes system message, windowed history, and current user message
func (m *Manager) BuildContextMessages(ctx context.Context, sessionID string, systemPrompt string, userMessage string) ([]*schema.Message, error) {
	msgs := make([]*schema.Message, 0)

	// Add system message if provided
	if systemPrompt != "" {
		msgs = append(msgs, schema.SystemMessage(systemPrompt))
	}

	// Add conversation history
	history, err := m.GetWindowedHistory(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	msgs = append(msgs, history...)

	// Add current user message
	if userMessage != "" {
		msgs = append(msgs, schema.UserMessage(userMessage))
	}

	return msgs, nil
}

// SummarizeIfNeeded checks if history exceeds threshold and summarizes
func (m *Manager) SummarizeIfNeeded(ctx context.Context, sessionID string) (bool, error) {
	if m.summarizer == nil {
		return false, nil
	}

	msgs, err := m.store.Read(ctx, sessionID)
	if err != nil {
		return false, err
	}

	summarized, didSummarize, err := m.summarizer.SummarizeIfNeeded(ctx, msgs)
	if err != nil {
		return false, err
	}

	if didSummarize {
		if err := m.store.Write(ctx, sessionID, summarized); err != nil {
			return false, err
		}
	}

	return didSummarize, nil
}

// GetStore returns the underlying store
func (m *Manager) GetStore() Store {
	return m.store
}
