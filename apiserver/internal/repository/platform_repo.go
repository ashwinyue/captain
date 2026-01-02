package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
)

type PlatformRepository struct {
	db *gorm.DB
}

func NewPlatformRepository(db *gorm.DB) *PlatformRepository {
	return &PlatformRepository{db: db}
}

// ListTypes returns all platform type definitions
func (r *PlatformRepository) ListTypes(ctx context.Context) ([]model.PlatformTypeDefinition, error) {
	var types []model.PlatformTypeDefinition
	err := r.db.WithContext(ctx).Order("name").Find(&types).Error
	return types, err
}

// List returns paginated platforms for a project
func (r *PlatformRepository) List(ctx context.Context, projectID uuid.UUID, platformType string, isActive *bool, limit, offset int) ([]model.Platform, int64, error) {
	var platforms []model.Platform
	var total int64

	query := r.db.WithContext(ctx).
		Preload("PlatformType").
		Where("project_id = ?", projectID)

	if platformType != "" {
		query = query.Where("type = ?", platformType)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Model(&model.Platform{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("is_active DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&platforms).Error

	return platforms, total, err
}

// GetByID returns a platform by ID
func (r *PlatformRepository) GetByID(ctx context.Context, projectID, platformID uuid.UUID) (*model.Platform, error) {
	var platform model.Platform
	err := r.db.WithContext(ctx).
		Preload("PlatformType").
		Where("id = ? AND project_id = ?", platformID, projectID).
		First(&platform).Error
	if err != nil {
		return nil, err
	}
	return &platform, nil
}

// GetByAPIKey returns a platform by API key
func (r *PlatformRepository) GetByAPIKey(ctx context.Context, apiKey string) (*model.Platform, error) {
	var platform model.Platform
	err := r.db.WithContext(ctx).
		Where("api_key = ?", apiKey).
		First(&platform).Error
	if err != nil {
		return nil, err
	}
	return &platform, nil
}

// Create creates a new platform
func (r *PlatformRepository) Create(ctx context.Context, platform *model.Platform) error {
	return r.db.WithContext(ctx).Create(platform).Error
}

// Update updates a platform
func (r *PlatformRepository) Update(ctx context.Context, platform *model.Platform) error {
	return r.db.WithContext(ctx).Save(platform).Error
}

// Delete soft deletes a platform
func (r *PlatformRepository) Delete(ctx context.Context, projectID, platformID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", platformID, projectID).
		Delete(&model.Platform{}).Error
}

// UpdateSyncStatus updates sync status fields
func (r *PlatformRepository) UpdateSyncStatus(ctx context.Context, platformID uuid.UUID, status string, err string) error {
	updates := map[string]interface{}{
		"sync_status": status,
		"sync_error":  err,
	}
	if status == string(model.PlatformSyncSynced) {
		updates["last_synced_at"] = gorm.Expr("NOW()")
	}
	return r.db.WithContext(ctx).
		Model(&model.Platform{}).
		Where("id = ?", platformID).
		Updates(updates).Error
}

// IncrementSyncRetry increments sync retry count
func (r *PlatformRepository) IncrementSyncRetry(ctx context.Context, platformID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Platform{}).
		Where("id = ?", platformID).
		UpdateColumn("sync_retry_count", gorm.Expr("sync_retry_count + 1")).Error
}
