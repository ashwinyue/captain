package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/rag/internal/model"
)

type EmbeddingConfigRepository struct {
	db *gorm.DB
}

func NewEmbeddingConfigRepository(db *gorm.DB) *EmbeddingConfigRepository {
	return &EmbeddingConfigRepository{db: db}
}

func (r *EmbeddingConfigRepository) FindAll(ctx context.Context) ([]model.EmbeddingConfig, error) {
	var configs []model.EmbeddingConfig
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Order("is_default DESC, name ASC").Find(&configs).Error
	return configs, err
}

func (r *EmbeddingConfigRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.EmbeddingConfig, error) {
	var config model.EmbeddingConfig
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *EmbeddingConfigRepository) FindDefault(ctx context.Context) (*model.EmbeddingConfig, error) {
	var config model.EmbeddingConfig
	err := r.db.WithContext(ctx).Where("is_default = true AND deleted_at IS NULL").First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *EmbeddingConfigRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID) (*model.EmbeddingConfig, error) {
	var config model.EmbeddingConfig
	err := r.db.WithContext(ctx).Where("project_id = ? AND deleted_at IS NULL", projectID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *EmbeddingConfigRepository) Upsert(ctx context.Context, config *model.EmbeddingConfig) error {
	var existing model.EmbeddingConfig
	err := r.db.WithContext(ctx).Where("project_id = ? AND deleted_at IS NULL", config.ProjectID).First(&existing).Error
	if err == nil {
		// Update existing
		existing.Provider = config.Provider
		existing.Model = config.Model
		existing.Dimensions = config.Dimensions
		existing.Name = config.Name
		return r.db.WithContext(ctx).Save(&existing).Error
	}
	// Create new
	return r.db.WithContext(ctx).Create(config).Error
}
