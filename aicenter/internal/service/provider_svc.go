package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/tgo/captain/aicenter/internal/model"
	"github.com/tgo/captain/aicenter/internal/repository"
)

type ProviderService struct {
	repo *repository.ProviderRepository
}

func NewProviderService(repo *repository.ProviderRepository) *ProviderService {
	return &ProviderService{repo: repo}
}

func (s *ProviderService) List(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.LLMProvider, int64, error) {
	return s.repo.List(ctx, projectID, limit, offset)
}

func (s *ProviderService) GetByID(ctx context.Context, projectID, providerID uuid.UUID) (*model.LLMProvider, error) {
	return s.repo.GetByID(ctx, projectID, providerID)
}

func (s *ProviderService) Create(ctx context.Context, provider *model.LLMProvider) error {
	return s.repo.Create(ctx, provider)
}

func (s *ProviderService) Update(ctx context.Context, provider *model.LLMProvider) error {
	return s.repo.Update(ctx, provider)
}

func (s *ProviderService) Delete(ctx context.Context, projectID, providerID uuid.UUID) error {
	return s.repo.Delete(ctx, projectID, providerID)
}

func (s *ProviderService) Sync(ctx context.Context, providers []model.LLMProvider) ([]model.LLMProvider, error) {
	result := make([]model.LLMProvider, 0, len(providers))
	for _, p := range providers {
		if err := s.repo.Upsert(ctx, &p); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}
