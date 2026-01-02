package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/tgo/captain/aicenter/internal/model"
	"github.com/tgo/captain/aicenter/internal/repository"
)

type TeamService struct {
	repo *repository.TeamRepository
}

func NewTeamService(repo *repository.TeamRepository) *TeamService {
	return &TeamService{repo: repo}
}

func (s *TeamService) List(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.Team, int64, error) {
	return s.repo.List(ctx, projectID, limit, offset)
}

func (s *TeamService) GetByID(ctx context.Context, projectID, teamID uuid.UUID) (*model.Team, error) {
	return s.repo.GetByID(ctx, projectID, teamID)
}

func (s *TeamService) GetWithAgents(ctx context.Context, projectID, teamID uuid.UUID) (*model.Team, error) {
	return s.repo.GetWithAgents(ctx, projectID, teamID)
}

func (s *TeamService) GetDefault(ctx context.Context, projectID uuid.UUID) (*model.Team, error) {
	return s.repo.GetDefault(ctx, projectID)
}

func (s *TeamService) GetOrCreateDefault(ctx context.Context, projectID uuid.UUID) (*model.Team, error) {
	team, err := s.repo.GetDefault(ctx, projectID)
	if err == nil {
		return team, nil
	}
	// Create default team if none exists
	defaultTeam := &model.Team{
		ProjectID:   projectID,
		Name:        "Default Team",
		Description: "Default team created automatically",
		IsDefault:   true,
	}
	if err := s.repo.Create(ctx, defaultTeam); err != nil {
		return nil, err
	}
	return defaultTeam, nil
}

func (s *TeamService) Create(ctx context.Context, team *model.Team) error {
	return s.repo.Create(ctx, team)
}

func (s *TeamService) Update(ctx context.Context, team *model.Team) error {
	return s.repo.Update(ctx, team)
}

func (s *TeamService) Delete(ctx context.Context, projectID, teamID uuid.UUID) error {
	return s.repo.Delete(ctx, projectID, teamID)
}
