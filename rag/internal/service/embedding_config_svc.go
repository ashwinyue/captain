package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/tgo/captain/rag/internal/model"
	"github.com/tgo/captain/rag/internal/repository"
)

type EmbeddingConfigService struct {
	repo *repository.EmbeddingConfigRepository
}

func NewEmbeddingConfigService(repo *repository.EmbeddingConfigRepository) *EmbeddingConfigService {
	return &EmbeddingConfigService{repo: repo}
}

func (s *EmbeddingConfigService) List(ctx context.Context) ([]model.EmbeddingConfig, error) {
	return s.repo.FindAll(ctx)
}

func (s *EmbeddingConfigService) GetByID(ctx context.Context, id uuid.UUID) (*model.EmbeddingConfig, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *EmbeddingConfigService) GetDefault(ctx context.Context) (*model.EmbeddingConfig, error) {
	return s.repo.FindDefault(ctx)
}

// GetByProjectID returns the active embedding config for a project
func (s *EmbeddingConfigService) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*model.EmbeddingConfig, error) {
	return s.repo.FindByProjectID(ctx, projectID)
}

// BatchSyncConfigRequest represents a single config in batch sync
type BatchSyncConfigRequest struct {
	ProjectID  uuid.UUID `json:"project_id"`
	Provider   string    `json:"provider"`
	Model      string    `json:"model"`
	Dimensions int       `json:"dimensions"`
	BatchSize  int       `json:"batch_size"`
	APIKey     string    `json:"api_key,omitempty"`
	BaseURL    string    `json:"base_url,omitempty"`
}

// BatchSyncResponse represents the batch sync response
type BatchSyncResponse struct {
	SuccessCount int                      `json:"success_count"`
	FailedCount  int                      `json:"failed_count"`
	Errors       []map[string]interface{} `json:"errors"`
}

// BatchSync upserts multiple embedding configs
func (s *EmbeddingConfigService) BatchSync(ctx context.Context, configs []BatchSyncConfigRequest) (*BatchSyncResponse, error) {
	successCount := 0
	var errors []map[string]interface{}

	for _, cfg := range configs {
		// Validate dimensions (phase 1 constraint)
		if cfg.Dimensions != 1536 {
			errors = append(errors, map[string]interface{}{
				"project_id": cfg.ProjectID.String(),
				"message":    "Invalid dimensions: only 1536 is supported in phase 1",
			})
			continue
		}

		// Try to upsert
		err := s.repo.Upsert(ctx, &model.EmbeddingConfig{
			ProjectID:  &cfg.ProjectID,
			Provider:   cfg.Provider,
			Model:      cfg.Model,
			Dimensions: cfg.Dimensions,
			Name:       cfg.Provider + "/" + cfg.Model,
			IsDefault:  false,
		})

		if err != nil {
			errors = append(errors, map[string]interface{}{
				"project_id": cfg.ProjectID.String(),
				"message":    err.Error(),
			})
			continue
		}

		successCount++
	}

	return &BatchSyncResponse{
		SuccessCount: successCount,
		FailedCount:  len(errors),
		Errors:       errors,
	}, nil
}
