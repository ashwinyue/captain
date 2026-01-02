package model

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/google/uuid"
)

type Agent struct {
	BaseModel
	ProjectID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	TeamID        *uuid.UUID `gorm:"type:uuid;index" json:"team_id,omitempty"`
	LLMProviderID *uuid.UUID `gorm:"type:uuid" json:"llm_provider_id,omitempty"`
	Name          string     `gorm:"size:255;not null" json:"name"`
	Description   string     `gorm:"type:text" json:"description"`
	Instruction   string     `gorm:"type:text" json:"instruction"`
	Model         string     `gorm:"size:255" json:"model"`
	IsDefault     bool       `gorm:"default:false" json:"is_default"`
	IsEnabled     bool       `gorm:"default:true" json:"is_enabled"`
	Config        JSONMap    `gorm:"type:jsonb" json:"config,omitempty"`

	// For runtime use (loaded separately, not via GORM relations)
	Team        *Team             `gorm:"-" json:"team,omitempty"`
	LLMProvider *LLMProvider      `gorm:"-" json:"llm_provider,omitempty"`
	Tools       []AgentTool       `gorm:"-" json:"tools,omitempty"`
	Collections []AgentCollection `gorm:"-" json:"collections,omitempty"`
}

func (Agent) TableName() string {
	return "ai_agents"
}

type AgentTool struct {
	BaseModel
	AgentID      uuid.UUID `gorm:"type:uuid;not null;index" json:"agent_id"`
	ToolProvider string    `gorm:"size:100;not null" json:"tool_provider"`
	ToolName     string    `gorm:"size:255;not null" json:"tool_name"`
	IsEnabled    bool      `gorm:"default:true" json:"is_enabled"`
	Config       JSONMap   `gorm:"type:jsonb" json:"config,omitempty"`
}

func (AgentTool) TableName() string {
	return "ai_agent_tools"
}

type AgentCollection struct {
	BaseModel
	AgentID      uuid.UUID `gorm:"type:uuid;not null;index" json:"agent_id"`
	CollectionID string    `gorm:"size:255;not null" json:"collection_id"`
	IsEnabled    bool      `gorm:"default:true" json:"is_enabled"`
}

// MarshalJSON customizes JSON serialization to return collection_id as id for frontend compatibility
func (ac AgentCollection) MarshalJSON() ([]byte, error) {
	type Alias AgentCollection
	return json.Marshal(&struct {
		ID string `json:"id"` // Use collection_id as id for frontend
		Alias
	}{
		ID:    ac.CollectionID,
		Alias: (Alias)(ac),
	})
}

func (AgentCollection) TableName() string {
	return "ai_agent_collections"
}

// JSONMap is a custom type for JSONB columns
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
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
