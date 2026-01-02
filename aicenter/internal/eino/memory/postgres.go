package memory

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConversationMessage represents a stored message in the database
type ConversationMessage struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SessionID string         `gorm:"size:255;index;not null"`
	ProjectID uuid.UUID      `gorm:"type:uuid;index"`
	UserID    string         `gorm:"size:255;index"`
	Role      string         `gorm:"size:50;not null"`
	Content   string         `gorm:"type:text"`
	Metadata  JSONMap        `gorm:"type:jsonb"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (ConversationMessage) TableName() string {
	return "conversation_messages"
}

// JSONMap for JSONB storage
type JSONMap map[string]interface{}

// Value implements driver.Valuer for GORM
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner for GORM
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}

	return json.Unmarshal(bytes, j)
}

// PostgresStore persists messages to PostgreSQL database
type PostgresStore struct {
	db        *gorm.DB
	projectID uuid.UUID
}

// NewPostgresStore creates a new PostgreSQL-backed store
func NewPostgresStore(db *gorm.DB, projectID uuid.UUID) *PostgresStore {
	return &PostgresStore{
		db:        db,
		projectID: projectID,
	}
}

// Write stores messages for the given session (replaces existing)
func (s *PostgresStore) Write(ctx context.Context, sessionID string, msgs []*schema.Message) error {
	// Delete existing messages for this session
	if err := s.Delete(ctx, sessionID); err != nil {
		return err
	}
	// Append new messages
	return s.Append(ctx, sessionID, msgs...)
}

// Read returns messages for the given session
func (s *PostgresStore) Read(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	var dbMsgs []ConversationMessage
	err := s.db.WithContext(ctx).
		Where("session_id = ? AND project_id = ?", sessionID, s.projectID).
		Order("created_at ASC").
		Find(&dbMsgs).Error
	if err != nil {
		return nil, err
	}

	msgs := make([]*schema.Message, 0, len(dbMsgs))
	for _, dbMsg := range dbMsgs {
		msg := &schema.Message{
			Role:    schema.RoleType(dbMsg.Role),
			Content: dbMsg.Content,
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

// Append adds messages to a session
func (s *PostgresStore) Append(ctx context.Context, sessionID string, msgs ...*schema.Message) error {
	if len(msgs) == 0 {
		return nil
	}

	dbMsgs := make([]ConversationMessage, 0, len(msgs))
	for _, msg := range msgs {
		if msg == nil {
			continue
		}

		metadata := make(JSONMap)
		if msg.ToolCalls != nil {
			toolCallsJSON, _ := json.Marshal(msg.ToolCalls)
			metadata["tool_calls"] = string(toolCallsJSON)
		}

		dbMsg := ConversationMessage{
			ID:        uuid.New(),
			SessionID: sessionID,
			ProjectID: s.projectID,
			Role:      string(msg.Role),
			Content:   msg.Content,
			Metadata:  metadata,
		}
		dbMsgs = append(dbMsgs, dbMsg)
	}

	return s.db.WithContext(ctx).Create(&dbMsgs).Error
}

// Delete removes a session's messages
func (s *PostgresStore) Delete(ctx context.Context, sessionID string) error {
	return s.db.WithContext(ctx).
		Where("session_id = ? AND project_id = ?", sessionID, s.projectID).
		Delete(&ConversationMessage{}).Error
}

// GetWindowedMessages returns the last N messages for a session
func (s *PostgresStore) GetWindowedMessages(ctx context.Context, sessionID string, windowSize int) ([]*schema.Message, error) {
	var dbMsgs []ConversationMessage
	err := s.db.WithContext(ctx).
		Where("session_id = ? AND project_id = ?", sessionID, s.projectID).
		Order("created_at DESC").
		Limit(windowSize).
		Find(&dbMsgs).Error
	if err != nil {
		return nil, err
	}

	// Reverse to get chronological order
	msgs := make([]*schema.Message, len(dbMsgs))
	for i, dbMsg := range dbMsgs {
		msgs[len(dbMsgs)-1-i] = &schema.Message{
			Role:    schema.RoleType(dbMsg.Role),
			Content: dbMsg.Content,
		}
	}
	return msgs, nil
}

// GetSessionCount returns the number of messages in a session
func (s *PostgresStore) GetSessionCount(ctx context.Context, sessionID string) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&ConversationMessage{}).
		Where("session_id = ? AND project_id = ?", sessionID, s.projectID).
		Count(&count).Error
	return count, err
}

// ListSessions returns all session IDs for the project
func (s *PostgresStore) ListSessions(ctx context.Context) ([]string, error) {
	var sessions []string
	err := s.db.WithContext(ctx).
		Model(&ConversationMessage{}).
		Where("project_id = ?", s.projectID).
		Distinct("session_id").
		Pluck("session_id", &sessions).Error
	return sessions, err
}
