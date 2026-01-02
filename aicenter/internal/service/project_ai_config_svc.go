package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/tgo/captain/aicenter/internal/model"
	"github.com/tgo/captain/aicenter/internal/repository"
)

type ProjectAIConfigService struct {
	repo *repository.ProjectAIConfigRepository
}

func NewProjectAIConfigService(repo *repository.ProjectAIConfigRepository) *ProjectAIConfigService {
	return &ProjectAIConfigService{repo: repo}
}

func (s *ProjectAIConfigService) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*model.ProjectAIConfig, error) {
	return s.repo.GetByProjectID(ctx, projectID)
}

func (s *ProjectAIConfigService) Upsert(ctx context.Context, config *model.ProjectAIConfig) error {
	return s.repo.Upsert(ctx, config)
}

func (s *ProjectAIConfigService) SyncConfigs(ctx context.Context, configs []*model.ProjectAIConfig) error {
	return s.repo.BulkUpsert(ctx, configs)
}

func (s *ProjectAIConfigService) Delete(ctx context.Context, projectID uuid.UUID) error {
	return s.repo.Delete(ctx, projectID)
}
