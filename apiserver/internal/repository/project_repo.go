package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
)

type ProjectRepository struct {
	BaseRepository[model.Project]
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{BaseRepository: BaseRepository[model.Project]{DB: db}}
}

func (r *ProjectRepository) FindByAPIKey(ctx context.Context, apiKey string) (*model.Project, error) {
	var project model.Project
	err := r.DB.WithContext(ctx).
		Where("api_key = ? AND is_active = true AND deleted_at IS NULL", apiKey).
		First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) UpdateAPIKey(ctx context.Context, id uuid.UUID, newKey string) error {
	return r.DB.WithContext(ctx).Model(&model.Project{}).
		Where("id = ?", id).
		Update("api_key", newKey).Error
}
