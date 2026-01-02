package model

import (
	"github.com/google/uuid"
)

type Team struct {
	BaseModel
	ProjectID             uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	SupervisorLLMID       *uuid.UUID `gorm:"type:uuid" json:"supervisor_llm_id,omitempty"`
	AIProviderID          *uuid.UUID `gorm:"type:uuid" json:"ai_provider_id,omitempty"`
	Name                  string     `gorm:"size:255;not null" json:"name"`
	Description           string     `gorm:"type:text" json:"description"`
	Model                 string     `gorm:"size:150" json:"model"`
	SupervisorInstruction string     `gorm:"type:text" json:"supervisor_instruction"`
	Instruction           string     `gorm:"type:text" json:"instruction"`
	ExpectedOutput        string     `gorm:"type:text" json:"expected_output"`
	SessionID             string     `gorm:"size:150" json:"session_id"`
	IsDefault             bool       `gorm:"default:false" json:"is_default"`
	IsEnabled             bool       `gorm:"default:true" json:"is_enabled"`
	Config                JSONMap    `gorm:"type:jsonb" json:"config,omitempty"`

	// For runtime use (populated via JOIN/Preload, not stored)
	SupervisorLLM *LLMProvider `gorm:"-" json:"-"`
	Agents        []Agent      `gorm:"-" json:"-"`
}

func (Team) TableName() string {
	return "ai_teams"
}
