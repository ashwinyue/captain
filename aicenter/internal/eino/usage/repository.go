package usage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles usage record persistence
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new usage repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new usage record
func (r *Repository) Create(ctx context.Context, record *UsageRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// GetByID retrieves a usage record by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*UsageRecord, error) {
	var record UsageRecord
	err := r.db.WithContext(ctx).First(&record, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListByProject retrieves usage records for a project
func (r *Repository) ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]UsageRecord, int64, error) {
	var records []UsageRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&UsageRecord{}).Where("project_id = ?", projectID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&records).Error
	return records, total, err
}

// ListBySession retrieves usage records for a session
func (r *Repository) ListBySession(ctx context.Context, projectID uuid.UUID, sessionID string) ([]UsageRecord, error) {
	var records []UsageRecord
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND session_id = ?", projectID, sessionID).
		Order("created_at ASC").
		Find(&records).Error
	return records, err
}

// GetSummary returns aggregated usage statistics for a project
func (r *Repository) GetSummary(ctx context.Context, projectID uuid.UUID, start, end time.Time) (*UsageSummary, error) {
	var result struct {
		TotalRequests         int64
		SuccessRequests       int64
		FailedRequests        int64
		TotalPromptTokens     int64
		TotalCompletionTokens int64
		TotalTokens           int64
		TotalCost             float64
		AvgLatencyMs          float64
	}

	err := r.db.WithContext(ctx).
		Model(&UsageRecord{}).
		Select(`
			COUNT(*) as total_requests,
			SUM(CASE WHEN success THEN 1 ELSE 0 END) as success_requests,
			SUM(CASE WHEN NOT success THEN 1 ELSE 0 END) as failed_requests,
			COALESCE(SUM(prompt_tokens), 0) as total_prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) as total_completion_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(cost), 0) as total_cost,
			COALESCE(AVG(latency_ms), 0) as avg_latency_ms
		`).
		Where("project_id = ? AND created_at >= ? AND created_at <= ?", projectID, start, end).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return &UsageSummary{
		ProjectID:             projectID,
		TotalRequests:         result.TotalRequests,
		SuccessRequests:       result.SuccessRequests,
		FailedRequests:        result.FailedRequests,
		TotalPromptTokens:     result.TotalPromptTokens,
		TotalCompletionTokens: result.TotalCompletionTokens,
		TotalTokens:           result.TotalTokens,
		TotalCost:             result.TotalCost,
		AvgLatencyMs:          result.AvgLatencyMs,
		PeriodStart:           start,
		PeriodEnd:             end,
	}, nil
}

// GetUsageByModel returns usage grouped by model
func (r *Repository) GetUsageByModel(ctx context.Context, projectID uuid.UUID, start, end time.Time) ([]UsageByModel, error) {
	var results []UsageByModel

	err := r.db.WithContext(ctx).
		Model(&UsageRecord{}).
		Select(`
			model,
			provider_kind,
			COUNT(*) as request_count,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(cost), 0) as total_cost
		`).
		Where("project_id = ? AND created_at >= ? AND created_at <= ?", projectID, start, end).
		Group("model, provider_kind").
		Order("total_tokens DESC").
		Scan(&results).Error

	return results, err
}

// GetUsageByAgent returns usage grouped by agent
func (r *Repository) GetUsageByAgent(ctx context.Context, projectID uuid.UUID, start, end time.Time) ([]UsageByAgent, error) {
	var results []UsageByAgent

	err := r.db.WithContext(ctx).
		Model(&UsageRecord{}).
		Select(`
			agent_id,
			COUNT(*) as request_count,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(cost), 0) as total_cost
		`).
		Where("project_id = ? AND agent_id IS NOT NULL AND created_at >= ? AND created_at <= ?", projectID, start, end).
		Group("agent_id").
		Order("total_tokens DESC").
		Scan(&results).Error

	return results, err
}

// DeleteOldRecords deletes records older than the specified time
func (r *Repository) DeleteOldRecords(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&UsageRecord{})
	return result.RowsAffected, result.Error
}
