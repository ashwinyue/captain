package model

import (
	"github.com/google/uuid"
)

// ToolType represents the type of tool
type ToolType string

const (
	ToolTypeMCP     ToolType = "mcp"
	ToolTypeRAG     ToolType = "rag"
	ToolTypeBuiltin ToolType = "builtin"
)

// TransportType represents the transport protocol for MCP tools
type TransportType string

const (
	TransportTypeHTTP  TransportType = "http"
	TransportTypeSSE   TransportType = "sse"
	TransportTypeStdio TransportType = "stdio"
)

// Tool represents a tool configuration
type Tool struct {
	BaseModel
	ProjectID     uuid.UUID     `gorm:"type:uuid;not null;index" json:"project_id"`
	Name          string        `gorm:"size:255;not null" json:"name"`
	Description   string        `gorm:"type:text" json:"description"`
	ToolType      ToolType      `gorm:"size:50;not null" json:"tool_type"`
	TransportType TransportType `gorm:"size:50" json:"transport_type,omitempty"`
	Endpoint      string        `gorm:"size:500" json:"endpoint,omitempty"`
	Config        JSONMap       `gorm:"type:jsonb" json:"config,omitempty"`
	IsEnabled     bool          `gorm:"default:true" json:"is_enabled"`
}

func (Tool) TableName() string {
	return "ai_tools"
}
