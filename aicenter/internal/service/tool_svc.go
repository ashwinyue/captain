package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/tgo/captain/aicenter/internal/model"
	"github.com/tgo/captain/aicenter/internal/repository"
)

type ToolService struct {
	repo *repository.ToolRepository
}

func NewToolService(repo *repository.ToolRepository) *ToolService {
	return &ToolService{repo: repo}
}

func (s *ToolService) List(ctx context.Context, projectID uuid.UUID, toolType *model.ToolType, includeDeleted bool, limit, offset int) ([]model.Tool, int64, error) {
	opts := &repository.ToolListOptions{
		ToolType:       toolType,
		IncludeDeleted: includeDeleted,
		Limit:          limit,
		Offset:         offset,
	}
	return s.repo.List(ctx, projectID, opts)
}

func (s *ToolService) GetByID(ctx context.Context, projectID, toolID uuid.UUID) (*model.Tool, error) {
	return s.repo.GetByID(ctx, projectID, toolID)
}

func (s *ToolService) Create(ctx context.Context, tool *model.Tool) error {
	return s.repo.Create(ctx, tool)
}

func (s *ToolService) Update(ctx context.Context, tool *model.Tool) error {
	return s.repo.Update(ctx, tool)
}

func (s *ToolService) Delete(ctx context.Context, projectID, toolID uuid.UUID) error {
	return s.repo.Delete(ctx, projectID, toolID)
}
