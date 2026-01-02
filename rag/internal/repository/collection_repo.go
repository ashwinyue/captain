package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/rag/internal/model"
)

type CollectionRepository struct {
	db *gorm.DB
}

func NewCollectionRepository(db *gorm.DB) *CollectionRepository {
	return &CollectionRepository{db: db}
}

func (r *CollectionRepository) Create(ctx context.Context, collection *model.Collection) error {
	return r.db.WithContext(ctx).Create(collection).Error
}

func (r *CollectionRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Collection, error) {
	var collection model.Collection
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&collection).Error
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

func (r *CollectionRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID, collectionType string, limit, offset int) ([]model.Collection, int64, error) {
	var collections []model.Collection
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Collection{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID)

	if collectionType != "" {
		query = query.Where("collection_type = ?", collectionType)
	}

	query.Count(&total)
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&collections).Error
	return collections, total, err
}

func (r *CollectionRepository) Update(ctx context.Context, collection *model.Collection) error {
	return r.db.WithContext(ctx).Save(collection).Error
}

func (r *CollectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Collection{}).Error
}
