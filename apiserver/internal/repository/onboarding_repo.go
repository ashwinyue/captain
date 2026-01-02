package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
)

type OnboardingRepository struct {
	db *gorm.DB
}

func NewOnboardingRepository(db *gorm.DB) *OnboardingRepository {
	return &OnboardingRepository{db: db}
}

// GetOrCreate returns the onboarding progress for a project, creating if not exists
func (r *OnboardingRepository) GetOrCreate(ctx context.Context, projectID uuid.UUID) (*model.ProjectOnboardingProgress, error) {
	var progress model.ProjectOnboardingProgress
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		First(&progress).Error

	if err == gorm.ErrRecordNotFound {
		progress = model.ProjectOnboardingProgress{
			ProjectID: projectID,
		}
		if err := r.db.WithContext(ctx).Create(&progress).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &progress, nil
}

// Update updates the onboarding progress
func (r *OnboardingRepository) Update(ctx context.Context, progress *model.ProjectOnboardingProgress) error {
	return r.db.WithContext(ctx).Save(progress).Error
}
