package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/repository"
)

type StaffService struct {
	repo *repository.StaffRepository
}

func NewStaffService(repo *repository.StaffRepository) *StaffService {
	return &StaffService{repo: repo}
}

func (s *StaffService) List(ctx context.Context, projectID *uuid.UUID, limit, offset int) ([]model.Staff, int64, error) {
	return s.repo.FindAllWithProject(ctx, projectID, limit, offset)
}

func (s *StaffService) GetByID(ctx context.Context, id uuid.UUID) (*model.Staff, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *StaffService) Create(ctx context.Context, staff *model.Staff) error {
	hash, err := HashPassword("changeme")
	if err != nil {
		return err
	}
	staff.PasswordHash = hash
	return s.repo.Create(ctx, staff)
}

func (s *StaffService) Update(ctx context.Context, staff *model.Staff) error {
	return s.repo.Update(ctx, staff)
}

func (s *StaffService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *StaffService) UpdateServicePaused(ctx context.Context, id uuid.UUID, paused bool) (*model.Staff, error) {
	staff, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	staff.ServicePaused = paused
	if err := s.repo.Update(ctx, staff); err != nil {
		return nil, err
	}
	return staff, nil
}

func (s *StaffService) UpdateIsActive(ctx context.Context, id uuid.UUID, isActive bool) (*model.Staff, error) {
	staff, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	staff.IsActive = isActive
	if err := s.repo.Update(ctx, staff); err != nil {
		return nil, err
	}
	return staff, nil
}
