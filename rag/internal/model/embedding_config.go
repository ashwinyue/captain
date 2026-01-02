package model

import (
	"github.com/google/uuid"
)

type EmbeddingConfig struct {
	BaseModel
	ProjectID      *uuid.UUID `gorm:"type:uuid;index" json:"project_id,omitempty"`
	Name           string     `gorm:"size:100;not null" json:"name"`
	Provider       string     `gorm:"size:50;not null" json:"provider"`
	Model          string     `gorm:"size:100;not null" json:"model"`
	Dimensions     int        `gorm:"not null" json:"dimensions"`
	MaxTokens      int        `gorm:"default:8191" json:"max_tokens"`
	IsDefault      bool       `gorm:"default:false" json:"is_default"`
	APIKeyRequired bool       `gorm:"default:true" json:"api_key_required"`
	Config         JSONMap    `gorm:"type:jsonb" json:"config"`
}

func (EmbeddingConfig) TableName() string {
	return "rag_embedding_configs"
}
