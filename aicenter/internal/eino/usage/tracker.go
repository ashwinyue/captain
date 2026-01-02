package usage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tracker tracks usage for LLM calls
type Tracker struct {
	repo *Repository
}

// NewTracker creates a new usage tracker
func NewTracker(db *gorm.DB) *Tracker {
	return &Tracker{
		repo: NewRepository(db),
	}
}

// TrackRequest represents a request to track
type TrackRequest struct {
	ProjectID        uuid.UUID
	AgentID          *uuid.UUID
	TeamID           *uuid.UUID
	SessionID        string
	RequestID        string
	ProviderKind     string
	Model            string
	PromptTokens     int
	CompletionTokens int
	LatencyMs        int64
	Success          bool
	ErrorMessage     string
	Metadata         map[string]interface{}
}

// Track records a usage event
func (t *Tracker) Track(ctx context.Context, req *TrackRequest) error {
	record := &UsageRecord{
		ID:               uuid.New(),
		ProjectID:        req.ProjectID,
		AgentID:          req.AgentID,
		TeamID:           req.TeamID,
		SessionID:        req.SessionID,
		RequestID:        req.RequestID,
		ProviderKind:     req.ProviderKind,
		Model:            req.Model,
		PromptTokens:     req.PromptTokens,
		CompletionTokens: req.CompletionTokens,
		TotalTokens:      req.PromptTokens + req.CompletionTokens,
		Cost:             calculateCost(req.ProviderKind, req.Model, req.PromptTokens, req.CompletionTokens),
		LatencyMs:        req.LatencyMs,
		Success:          req.Success,
		ErrorMessage:     req.ErrorMessage,
		Metadata:         JSONMap(req.Metadata),
		CreatedAt:        time.Now(),
	}

	return t.repo.Create(ctx, record)
}

// GetSummary returns usage summary for a project
func (t *Tracker) GetSummary(ctx context.Context, projectID uuid.UUID, start, end time.Time) (*UsageSummary, error) {
	return t.repo.GetSummary(ctx, projectID, start, end)
}

// GetDailySummary returns usage summary for today
func (t *Tracker) GetDailySummary(ctx context.Context, projectID uuid.UUID) (*UsageSummary, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	return t.repo.GetSummary(ctx, projectID, start, end)
}

// GetMonthlySummary returns usage summary for the current month
func (t *Tracker) GetMonthlySummary(ctx context.Context, projectID uuid.UUID) (*UsageSummary, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0)
	return t.repo.GetSummary(ctx, projectID, start, end)
}

// GetUsageByModel returns usage breakdown by model
func (t *Tracker) GetUsageByModel(ctx context.Context, projectID uuid.UUID, start, end time.Time) ([]UsageByModel, error) {
	return t.repo.GetUsageByModel(ctx, projectID, start, end)
}

// GetUsageByAgent returns usage breakdown by agent
func (t *Tracker) GetUsageByAgent(ctx context.Context, projectID uuid.UUID, start, end time.Time) ([]UsageByAgent, error) {
	return t.repo.GetUsageByAgent(ctx, projectID, start, end)
}

// calculateCost calculates the cost based on provider pricing
func calculateCost(providerKind, model string, promptTokens, completionTokens int) float64 {
	// Pricing per 1M tokens (simplified)
	var promptPrice, completionPrice float64

	switch providerKind {
	case "openai":
		switch model {
		case "gpt-4", "gpt-4-turbo":
			promptPrice = 30.0 // $30 per 1M tokens
			completionPrice = 60.0
		case "gpt-4o":
			promptPrice = 5.0
			completionPrice = 15.0
		case "gpt-4o-mini":
			promptPrice = 0.15
			completionPrice = 0.60
		case "gpt-3.5-turbo":
			promptPrice = 0.5
			completionPrice = 1.5
		default:
			promptPrice = 1.0
			completionPrice = 2.0
		}
	case "anthropic":
		switch model {
		case "claude-3-opus":
			promptPrice = 15.0
			completionPrice = 75.0
		case "claude-3-sonnet":
			promptPrice = 3.0
			completionPrice = 15.0
		case "claude-3-haiku":
			promptPrice = 0.25
			completionPrice = 1.25
		default:
			promptPrice = 3.0
			completionPrice = 15.0
		}
	default:
		promptPrice = 1.0
		completionPrice = 2.0
	}

	// Calculate cost (prices are per 1M tokens)
	cost := (float64(promptTokens) * promptPrice / 1000000) +
		(float64(completionTokens) * completionPrice / 1000000)

	return cost
}
