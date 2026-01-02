package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/tgo/captain/aicenter/internal/model"
)

type ProjectAIConfigRepository struct {
	db *gorm.DB
}

func NewProjectAIConfigRepository(db *gorm.DB) *ProjectAIConfigRepository {
	return &ProjectAIConfigRepository{db: db}
}

func (r *ProjectAIConfigRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*model.ProjectAIConfig, error) {
	var config model.ProjectAIConfig
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *ProjectAIConfigRepository) Upsert(ctx context.Context, config *model.ProjectAIConfig) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "project_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"default_llm_provider_id", "default_model", "default_embedding_provider_id", "default_embedding_model", "default_team_id", "config", "updated_at"}),
		}).
		Create(config).Error
}

func (r *ProjectAIConfigRepository) BulkUpsert(ctx context.Context, configs []*model.ProjectAIConfig) error {
	if len(configs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, config := range configs {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "project_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"default_llm_provider_id", "default_model", "default_embedding_provider_id", "default_embedding_model", "default_team_id", "config", "updated_at"}),
			}).Create(config).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ProjectAIConfigRepository) Delete(ctx context.Context, projectID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Delete(&model.ProjectAIConfig{}).Error
}
