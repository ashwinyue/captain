package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/rag/internal/model"
)

type QAPairRepository struct {
	db *gorm.DB
}

func NewQAPairRepository(db *gorm.DB) *QAPairRepository {
	return &QAPairRepository{db: db}
}

func (r *QAPairRepository) Create(ctx context.Context, qa *model.QAPair) error {
	return r.db.WithContext(ctx).Create(qa).Error
}

func (r *QAPairRepository) CreateBatch(ctx context.Context, qas []model.QAPair) error {
	return r.db.WithContext(ctx).Create(&qas).Error
}

func (r *QAPairRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.QAPair, error) {
	var qa model.QAPair
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&qa).Error
	if err != nil {
		return nil, err
	}
	return &qa, nil
}

func (r *QAPairRepository) FindByCollectionID(ctx context.Context, collectionID uuid.UUID, category, status string, limit, offset int) ([]model.QAPair, int64, error) {
	var qas []model.QAPair
	var total int64

	query := r.db.WithContext(ctx).Model(&model.QAPair{}).
		Where("collection_id = ? AND deleted_at IS NULL", collectionID)

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.Order("priority DESC, created_at DESC").Limit(limit).Offset(offset).Find(&qas).Error
	return qas, total, err
}

func (r *QAPairRepository) Update(ctx context.Context, qa *model.QAPair) error {
	return r.db.WithContext(ctx).Save(qa).Error
}

func (r *QAPairRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.QAPair{}).Error
}

func (r *QAPairRepository) ListCategories(ctx context.Context, projectID uuid.UUID, collectionID *uuid.UUID) ([]string, error) {
	var categories []string

	query := r.db.WithContext(ctx).Model(&model.QAPair{}).
		Select("DISTINCT category").
		Where("project_id = ? AND deleted_at IS NULL AND category IS NOT NULL AND category != ''", projectID)

	if collectionID != nil {
		query = query.Where("collection_id = ?", *collectionID)
	}

	err := query.Pluck("category", &categories).Error
	return categories, err
}

// QAStats represents statistics for QA pairs
type QAStats struct {
	Total     int64 `json:"total"`
	Pending   int64 `json:"pending"`
	Processed int64 `json:"processed"`
	Failed    int64 `json:"failed"`
}

// GetStats returns QA pair statistics for a collection
func (r *QAPairRepository) GetStats(ctx context.Context, collectionID uuid.UUID) (*QAStats, error) {
	var stats QAStats

	// Total count
	err := r.db.WithContext(ctx).Model(&model.QAPair{}).
		Where("collection_id = ? AND deleted_at IS NULL", collectionID).
		Count(&stats.Total).Error
	if err != nil {
		return nil, err
	}

	// Pending (draft or inactive without vector_id)
	err = r.db.WithContext(ctx).Model(&model.QAPair{}).
		Where("collection_id = ? AND deleted_at IS NULL AND (status = 'draft' OR (vector_id IS NULL OR vector_id = ''))", collectionID).
		Count(&stats.Pending).Error
	if err != nil {
		return nil, err
	}

	// Processed (active with vector_id)
	err = r.db.WithContext(ctx).Model(&model.QAPair{}).
		Where("collection_id = ? AND deleted_at IS NULL AND status = 'active' AND vector_id IS NOT NULL AND vector_id != ''", collectionID).
		Count(&stats.Processed).Error
	if err != nil {
		return nil, err
	}

	// Failed (inactive status)
	err = r.db.WithContext(ctx).Model(&model.QAPair{}).
		Where("collection_id = ? AND deleted_at IS NULL AND status = 'inactive'", collectionID).
		Count(&stats.Failed).Error
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// CountByCollection returns the total count of QA pairs in a collection
func (r *QAPairRepository) CountByCollection(ctx context.Context, collectionID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.QAPair{}).
		Where("collection_id = ? AND deleted_at IS NULL", collectionID).
		Count(&count).Error
	return count, err
}
