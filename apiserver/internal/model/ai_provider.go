package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AIProvider represents an AI/LLM provider configuration
type AIProvider struct {
	ID              uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID       uuid.UUID   `gorm:"type:uuid;not null" json:"project_id"`
	Provider        string      `gorm:"size:50;not null" json:"provider"`
	Name            string      `gorm:"size:100;not null" json:"name"`
	APIKey          string      `gorm:"size:500" json:"api_key,omitempty"`
	APIBaseURL      string      `gorm:"size:500" json:"api_base_url,omitempty"`
	DefaultModel    string      `gorm:"size:100" json:"default_model,omitempty"`
	AvailableModels StringArray `gorm:"type:jsonb;default:'[]'" json:"available_models,omitempty"`
	Config          JSONMap     `gorm:"type:jsonb;default:'{}'" json:"config,omitempty"`
	IsActive        bool        `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       *time.Time  `gorm:"index" json:"deleted_at,omitempty"`
}

func (AIProvider) TableName() string {
	return "api_ai_providers"
}

// StringArray is a custom type for JSON string arrays with GORM support
type StringArray []string

// Value implements driver.Valuer
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Scan implements sql.Scanner
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, a)
}
