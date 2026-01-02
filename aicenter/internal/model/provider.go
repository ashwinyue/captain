package model

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/google/uuid"
)

type LLMProvider struct {
	BaseModel
	ProjectID       uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	Name            string    `gorm:"size:255;not null" json:"name"`
	ProviderKind    string    `gorm:"column:provider_kind;size:50;not null" json:"provider"` // JSON uses "provider" to match Python API
	Vendor          string    `gorm:"size:100" json:"vendor,omitempty"`
	APIKey          string    `gorm:"size:500" json:"api_key,omitempty"` // 创建时接收，响应时不返回
	APIBaseURL      string    `gorm:"size:500" json:"api_base_url,omitempty"`
	DefaultModel    string    `gorm:"column:model;size:255" json:"default_model,omitempty"`
	AvailableModels JSONArray `gorm:"type:jsonb" json:"available_models,omitempty"`
	Organization    string    `gorm:"size:255" json:"organization,omitempty"`
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	Timeout         float64   `gorm:"default:60" json:"timeout"`
	Config          JSONMap   `gorm:"type:jsonb" json:"config,omitempty"`
}

// JSONArray is a custom type for JSONB array
type JSONArray []string

func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (LLMProvider) TableName() string {
	return "ai_llm_providers"
}
