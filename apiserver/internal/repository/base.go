package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseRepository[T any] struct {
	DB *gorm.DB
}

func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Create(entity).Error
}

func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Save(entity).Error
}

func (r *BaseRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	var entity T
	return r.DB.WithContext(ctx).Where("id = ?", id).Delete(&entity).Error
}

func (r *BaseRepository[T]) FindByID(ctx context.Context, id uuid.UUID) (*T, error) {
	var entity T
	err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *BaseRepository[T]) FindAll(ctx context.Context, limit, offset int) ([]T, int64, error) {
	var entities []T
	var total int64

	query := r.DB.WithContext(ctx).Model(new(T)).Where("deleted_at IS NULL")
	query.Count(&total)
	err := query.Limit(limit).Offset(offset).Find(&entities).Error
	return entities, total, err
}
