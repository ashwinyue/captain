package model

import (
	"github.com/google/uuid"
)

// ProjectAIConfig stores project-level default AI model configurations
type ProjectAIConfig struct {
	BaseModel
	ProjectID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"project_id"`

	// Chat model settings (use "default_chat_*" in JSON to match Python API)
	DefaultChatProviderID *uuid.UUID `gorm:"column:default_llm_provider_id;type:uuid" json:"default_chat_provider_id,omitempty"`
	DefaultChatModel      string     `gorm:"column:default_model;size:255" json:"default_chat_model,omitempty"`

	// Embedding settings
	DefaultEmbeddingProviderID *uuid.UUID `gorm:"type:uuid" json:"default_embedding_provider_id,omitempty"`
	DefaultEmbeddingModel      string     `gorm:"size:255" json:"default_embedding_model,omitempty"`

	// Other settings
	DefaultTeamID *uuid.UUID `gorm:"type:uuid" json:"default_team_id,omitempty"`
	Config        JSONMap    `gorm:"type:jsonb" json:"config,omitempty"`

	// Sync status
	SyncStatus       string `gorm:"size:50" json:"sync_status,omitempty"`
	SyncError        string `gorm:"type:text" json:"sync_error,omitempty"`
	SyncAttemptCount int    `gorm:"default:0" json:"sync_attempt_count"`
	LastSyncAt       *int64 `json:"last_sync_at,omitempty"`

	// For runtime use (loaded separately, not via GORM relations)
	DefaultChatProvider      *LLMProvider `gorm:"-" json:"default_chat_provider,omitempty"`
	DefaultTeam              *Team        `gorm:"-" json:"default_team,omitempty"`
	DefaultEmbeddingProvider *LLMProvider `gorm:"-" json:"default_embedding_provider,omitempty"`
}

func (ProjectAIConfig) TableName() string {
	return "ai_project_configs"
}
