package usage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UsageRecord represents a single usage record for token tracking
type UsageRecord struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID        uuid.UUID      `gorm:"type:uuid;index;not null"`
	AgentID          *uuid.UUID     `gorm:"type:uuid;index"`
	TeamID           *uuid.UUID     `gorm:"type:uuid;index"`
	SessionID        string         `gorm:"size:255;index"`
	RequestID        string         `gorm:"size:255;index"`
	ProviderKind     string         `gorm:"size:50"`
	Model            string         `gorm:"size:100"`
	PromptTokens     int            `gorm:"default:0"`
	CompletionTokens int            `gorm:"default:0"`
	TotalTokens      int            `gorm:"default:0"`
	Cost             float64        `gorm:"type:decimal(10,6);default:0"`
	LatencyMs        int64          `gorm:"default:0"`
	Success          bool           `gorm:"default:true"`
	ErrorMessage     string         `gorm:"type:text"`
	Metadata         JSONMap        `gorm:"type:jsonb"`
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

func (UsageRecord) TableName() string {
	return "usage_records"
}

// JSONMap for JSONB storage
type JSONMap map[string]interface{}

// UsageSummary represents aggregated usage statistics
type UsageSummary struct {
	ProjectID             uuid.UUID `json:"project_id"`
	TotalRequests         int64     `json:"total_requests"`
	SuccessRequests       int64     `json:"success_requests"`
	FailedRequests        int64     `json:"failed_requests"`
	TotalPromptTokens     int64     `json:"total_prompt_tokens"`
	TotalCompletionTokens int64     `json:"total_completion_tokens"`
	TotalTokens           int64     `json:"total_tokens"`
	TotalCost             float64   `json:"total_cost"`
	AvgLatencyMs          float64   `json:"avg_latency_ms"`
	PeriodStart           time.Time `json:"period_start"`
	PeriodEnd             time.Time `json:"period_end"`
}

// UsageByModel represents usage grouped by model
type UsageByModel struct {
	Model        string  `json:"model"`
	ProviderKind string  `json:"provider_kind"`
	RequestCount int64   `json:"request_count"`
	TotalTokens  int64   `json:"total_tokens"`
	TotalCost    float64 `json:"total_cost"`
}

// UsageByAgent represents usage grouped by agent
type UsageByAgent struct {
	AgentID      uuid.UUID `json:"agent_id"`
	AgentName    string    `json:"agent_name"`
	RequestCount int64     `json:"request_count"`
	TotalTokens  int64     `json:"total_tokens"`
	TotalCost    float64   `json:"total_cost"`
}
