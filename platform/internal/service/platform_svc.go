package service

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/platform/internal/model"
)

type PlatformService struct {
	db *gorm.DB
}

func NewPlatformService(db *gorm.DB) *PlatformService {
	return &PlatformService{db: db}
}

func (s *PlatformService) List(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.Platform, int64, error) {
	var platforms []model.Platform
	var total int64

	query := s.db.WithContext(ctx).Model(&model.Platform{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID)

	query.Count(&total)
	err := query.Limit(limit).Offset(offset).Find(&platforms).Error
	return platforms, total, err
}

func (s *PlatformService) GetByID(ctx context.Context, projectID, platformID uuid.UUID) (*model.Platform, error) {
	var platform model.Platform
	err := s.db.WithContext(ctx).
		Where("project_id = ? AND id = ? AND deleted_at IS NULL", projectID, platformID).
		First(&platform).Error
	return &platform, err
}

func (s *PlatformService) Create(ctx context.Context, platform *model.Platform) error {
	return s.db.WithContext(ctx).Create(platform).Error
}

func (s *PlatformService) Update(ctx context.Context, platform *model.Platform) error {
	return s.db.WithContext(ctx).Save(platform).Error
}

func (s *PlatformService) Delete(ctx context.Context, projectID, platformID uuid.UUID) error {
	return s.db.WithContext(ctx).
		Where("project_id = ? AND id = ?", projectID, platformID).
		Delete(&model.Platform{}).Error
}

func (s *PlatformService) Enable(ctx context.Context, projectID, platformID uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&model.Platform{}).
		Where("project_id = ? AND id = ?", projectID, platformID).
		Update("is_enabled", true).Error
}

func (s *PlatformService) Disable(ctx context.Context, projectID, platformID uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&model.Platform{}).
		Where("project_id = ? AND id = ?", projectID, platformID).
		Update("is_enabled", false).Error
}
