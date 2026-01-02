package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/aicenter/internal/model"
)

type ProviderRepository struct {
	db *gorm.DB
}

func NewProviderRepository(db *gorm.DB) *ProviderRepository {
	return &ProviderRepository{db: db}
}

func (r *ProviderRepository) List(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.LLMProvider, int64, error) {
	var providers []model.LLMProvider
	var total int64

	query := r.db.WithContext(ctx).Where("project_id = ?", projectID)

	if err := query.Model(&model.LLMProvider{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&providers).Error; err != nil {
		return nil, 0, err
	}

	return providers, total, nil
}

func (r *ProviderRepository) GetByID(ctx context.Context, projectID, providerID uuid.UUID) (*model.LLMProvider, error) {
	var provider model.LLMProvider
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND id = ?", projectID, providerID).
		First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *ProviderRepository) Create(ctx context.Context, provider *model.LLMProvider) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

func (r *ProviderRepository) Update(ctx context.Context, provider *model.LLMProvider) error {
	return r.db.WithContext(ctx).Save(provider).Error
}

func (r *ProviderRepository) Delete(ctx context.Context, projectID, providerID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND id = ?", projectID, providerID).
		Delete(&model.LLMProvider{}).Error
}

func (r *ProviderRepository) Upsert(ctx context.Context, provider *model.LLMProvider) error {
	return r.db.WithContext(ctx).Save(provider).Error
}

func (r *ProviderRepository) GetByProjectAndAlias(ctx context.Context, projectID uuid.UUID, alias string) (*model.LLMProvider, error) {
	var provider model.LLMProvider
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND alias = ?", projectID, alias).
		First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}
