package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/repository"
)

type ProjectService struct {
	repo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) List(ctx context.Context, limit, offset int) ([]model.Project, int64, error) {
	return s.repo.FindAll(ctx, limit, offset)
}

func (s *ProjectService) GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ProjectService) Create(ctx context.Context, project *model.Project) error {
	if project.APIKey == "" {
		project.APIKey = generateAPIKey()
	}
	return s.repo.Create(ctx, project)
}

func (s *ProjectService) Update(ctx context.Context, project *model.Project) error {
	return s.repo.Update(ctx, project)
}

func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *ProjectService) RegenerateAPIKey(ctx context.Context, id uuid.UUID) (string, error) {
	newKey := generateAPIKey()
	err := s.repo.UpdateAPIKey(ctx, id, newKey)
	return newKey, err
}

func generateAPIKey() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "ak_" + hex.EncodeToString(bytes)
}
