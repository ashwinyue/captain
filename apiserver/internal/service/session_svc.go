package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/repository"
)

type SessionService struct {
	repo *repository.SessionRepository
}

func NewSessionService(repo *repository.SessionRepository) *SessionService {
	return &SessionService{repo: repo}
}

func (s *SessionService) List(ctx context.Context, projectID uuid.UUID, status *model.SessionStatus, limit, offset int) ([]model.Session, int64, error) {
	return s.repo.FindByProjectID(ctx, projectID, status, limit, offset)
}

func (s *SessionService) GetByID(ctx context.Context, projectID, sessionID uuid.UUID) (*model.Session, error) {
	return s.repo.FindByIDAndProject(ctx, projectID, sessionID)
}

func (s *SessionService) Create(ctx context.Context, session *model.Session) error {
	return s.repo.Create(ctx, session)
}

func (s *SessionService) Update(ctx context.Context, session *model.Session) error {
	return s.repo.Update(ctx, session)
}

func (s *SessionService) Close(ctx context.Context, projectID, sessionID uuid.UUID) error {
	return s.repo.Close(ctx, projectID, sessionID)
}

func (s *SessionService) Transfer(ctx context.Context, projectID, sessionID uuid.UUID, toStaffID uuid.UUID) error {
	return s.repo.Transfer(ctx, projectID, sessionID, toStaffID)
}
