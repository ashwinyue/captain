package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/rag/internal/model"
)

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, file *model.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *FileRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.File, error) {
	var file model.File
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID, collectionID *uuid.UUID, status string, limit, offset int) ([]model.File, int64, error) {
	var files []model.File
	var total int64

	query := r.db.WithContext(ctx).Model(&model.File{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID)

	if collectionID != nil {
		query = query.Where("collection_id = ?", *collectionID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&files).Error
	return files, total, err
}

func (r *FileRepository) Update(ctx context.Context, file *model.File) error {
	return r.db.WithContext(ctx).Save(file).Error
}

func (r *FileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.File{}).Error
}
