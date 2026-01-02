package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/tgo/captain/platform/internal/model"
)

type OnboardingService struct {
	db *gorm.DB
}

func NewOnboardingService(db *gorm.DB) *OnboardingService {
	return &OnboardingService{db: db}
}

func (s *OnboardingService) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*model.Onboarding, error) {
	var onboarding model.Onboarding
	err := s.db.WithContext(ctx).
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		First(&onboarding).Error
	if err == gorm.ErrRecordNotFound {
		// Create default onboarding
		onboarding = model.Onboarding{
			ProjectID:  projectID,
			TotalSteps: 5,
		}
		s.db.WithContext(ctx).Create(&onboarding)
		return &onboarding, nil
	}
	return &onboarding, err
}

func (s *OnboardingService) UpdateStep(ctx context.Context, projectID uuid.UUID, step int, completed bool) error {
	onboarding, err := s.GetByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	if onboarding.CompletedSteps == nil {
		onboarding.CompletedSteps = make(model.JSONMap)
	}

	if completed {
		onboarding.CompletedSteps[string(rune('0'+step))] = true
		if step >= onboarding.CurrentStep {
			onboarding.CurrentStep = step + 1
		}
	}

	// Check if all steps completed
	completedCount := 0
	for range onboarding.CompletedSteps {
		completedCount++
	}
	if completedCount >= onboarding.TotalSteps {
		onboarding.IsCompleted = true
		now := time.Now()
		onboarding.CompletedAt = &now
	}

	return s.db.WithContext(ctx).Save(onboarding).Error
}

func (s *OnboardingService) SkipStep(ctx context.Context, projectID uuid.UUID, step int) error {
	onboarding, err := s.GetByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	if onboarding.SkippedSteps == nil {
		onboarding.SkippedSteps = make(model.JSONMap)
	}
	onboarding.SkippedSteps[string(rune('0'+step))] = true

	if step >= onboarding.CurrentStep {
		onboarding.CurrentStep = step + 1
	}

	return s.db.WithContext(ctx).Save(onboarding).Error
}

func (s *OnboardingService) Reset(ctx context.Context, projectID uuid.UUID) error {
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "project_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"current_step", "is_completed", "completed_steps", "skipped_steps", "completed_at"}),
		}).
		Create(&model.Onboarding{
			ProjectID:  projectID,
			TotalSteps: 5,
		}).Error
}
