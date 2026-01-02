package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
)

type VisitorRepository struct {
	BaseRepository[model.Visitor]
}

func NewVisitorRepository(db *gorm.DB) *VisitorRepository {
	return &VisitorRepository{BaseRepository: BaseRepository[model.Visitor]{DB: db}}
}

func (r *VisitorRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.Visitor, int64, error) {
	var visitors []model.Visitor
	var total int64

	query := r.DB.WithContext(ctx).Model(&model.Visitor{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID)

	query.Count(&total)
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&visitors).Error
	return visitors, total, err
}

func (r *VisitorRepository) FindByIDAndProject(ctx context.Context, projectID, id uuid.UUID) (*model.Visitor, error) {
	var visitor model.Visitor
	err := r.DB.WithContext(ctx).
		Where("id = ? AND project_id = ? AND deleted_at IS NULL", id, projectID).
		First(&visitor).Error
	if err != nil {
		return nil, err
	}
	return &visitor, nil
}

func (r *VisitorRepository) UpdateBlocked(ctx context.Context, projectID, id uuid.UUID, blocked bool) error {
	return r.DB.WithContext(ctx).Model(&model.Visitor{}).
		Where("id = ? AND project_id = ?", id, projectID).
		Update("is_blocked", blocked).Error
}

func (r *VisitorRepository) Search(ctx context.Context, projectID uuid.UUID, keyword string, limit, offset int) ([]model.Visitor, int64, error) {
	var visitors []model.Visitor
	var total int64

	searchTerm := "%" + keyword + "%"
	query := r.DB.WithContext(ctx).Model(&model.Visitor{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Where("name ILIKE ? OR email ILIKE ? OR phone ILIKE ? OR external_id ILIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm)

	query.Count(&total)
	err := query.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&visitors).Error
	return visitors, total, err
}

func (r *VisitorRepository) FindByPlatformOpenID(ctx context.Context, projectID, platformID uuid.UUID, platformOpenID string) (*model.Visitor, error) {
	var visitor model.Visitor
	err := r.DB.WithContext(ctx).
		Where("project_id = ? AND platform_id = ? AND platform_open_id = ? AND deleted_at IS NULL",
			projectID, platformID, platformOpenID).
		First(&visitor).Error
	if err != nil {
		return nil, err
	}
	return &visitor, nil
}

// GetPlatformByAPIKey gets platform by API key
func (r *VisitorRepository) GetPlatformByAPIKey(ctx context.Context, apiKey string) (*model.Platform, error) {
	var platform model.Platform
	err := r.DB.WithContext(ctx).
		Where("api_key = ? AND is_active = true AND deleted_at IS NULL", apiKey).
		First(&platform).Error
	if err != nil {
		return nil, err
	}
	return &platform, nil
}

// FindByServiceStatus finds all visitors with a specific service status
func (r *VisitorRepository) FindByServiceStatus(ctx context.Context, status string) ([]model.Visitor, error) {
	var visitors []model.Visitor
	err := r.DB.WithContext(ctx).
		Where("service_status = ? AND deleted_at IS NULL", status).
		Find(&visitors).Error
	return visitors, err
}

// ResetToAIMode resets a visitor to AI mode
func (r *VisitorRepository) ResetToAIMode(ctx context.Context, visitorID uuid.UUID) error {
	return r.DB.WithContext(ctx).Model(&model.Visitor{}).
		Where("id = ?", visitorID).
		Updates(map[string]interface{}{
			"service_status":    "new",
			"ai_enabled":        true,
			"assigned_staff_id": nil,
		}).Error
}
