package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/tgo/captain/aicenter/internal/model"
	"github.com/tgo/captain/aicenter/internal/repository"
)

type AgentService struct {
	repo *repository.AgentRepository
}

func NewAgentService(repo *repository.AgentRepository) *AgentService {
	return &AgentService{repo: repo}
}

func (s *AgentService) List(ctx context.Context, projectID uuid.UUID, teamID *uuid.UUID, limit, offset int) ([]model.Agent, int64, error) {
	opts := []repository.ListOption{
		repository.WithPagination(limit, offset),
	}
	if teamID != nil {
		opts = append(opts, repository.WithTeamID(*teamID))
	}
	return s.repo.List(ctx, projectID, opts...)
}

func (s *AgentService) GetByID(ctx context.Context, projectID, agentID uuid.UUID) (*model.Agent, error) {
	return s.repo.GetByID(ctx, projectID, agentID)
}

func (s *AgentService) Create(ctx context.Context, agent *model.Agent) error {
	return s.repo.Create(ctx, agent)
}

func (s *AgentService) Update(ctx context.Context, agent *model.Agent) error {
	return s.repo.Update(ctx, agent)
}

func (s *AgentService) Delete(ctx context.Context, projectID, agentID uuid.UUID) error {
	return s.repo.Delete(ctx, projectID, agentID)
}

func (s *AgentService) Exists(ctx context.Context, projectID uuid.UUID) (bool, int64, error) {
	return s.repo.Exists(ctx, projectID)
}

func (s *AgentService) SetToolEnabled(ctx context.Context, projectID, agentID, toolID uuid.UUID, enabled bool) error {
	return s.repo.SetToolEnabled(ctx, projectID, agentID, toolID, enabled)
}

func (s *AgentService) SetCollectionEnabled(ctx context.Context, projectID, agentID uuid.UUID, collectionID string, enabled bool) error {
	return s.repo.SetCollectionEnabled(ctx, projectID, agentID, collectionID, enabled)
}
