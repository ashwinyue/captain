package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/rag/internal/model"
)

type WebsitePageRepository struct {
	db *gorm.DB
}

func NewWebsitePageRepository(db *gorm.DB) *WebsitePageRepository {
	return &WebsitePageRepository{db: db}
}

func (r *WebsitePageRepository) Create(ctx context.Context, page *model.WebsitePage) error {
	return r.db.WithContext(ctx).Create(page).Error
}

func (r *WebsitePageRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.WebsitePage, error) {
	var page model.WebsitePage
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&page).Error
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *WebsitePageRepository) FindByCollectionID(ctx context.Context, collectionID uuid.UUID, status string, limit, offset int) ([]model.WebsitePage, int64, error) {
	var pages []model.WebsitePage
	var total int64

	query := r.db.WithContext(ctx).Model(&model.WebsitePage{}).
		Where("collection_id = ? AND deleted_at IS NULL", collectionID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.Order("depth ASC, created_at DESC").Limit(limit).Offset(offset).Find(&pages).Error
	return pages, total, err
}

func (r *WebsitePageRepository) Update(ctx context.Context, page *model.WebsitePage) error {
	return r.db.WithContext(ctx).Save(page).Error
}

func (r *WebsitePageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.WebsitePage{}).Error
}

func (r *WebsitePageRepository) GetCrawlProgress(ctx context.Context, collectionID uuid.UUID) (map[string]int64, error) {
	progress := make(map[string]int64)

	var results []struct {
		Status model.PageStatus
		Count  int64
	}

	err := r.db.WithContext(ctx).Model(&model.WebsitePage{}).
		Select("status, count(*) as count").
		Where("collection_id = ? AND deleted_at IS NULL", collectionID).
		Group("status").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	for _, r := range results {
		progress[string(r.Status)] = r.Count
	}

	return progress, nil
}
