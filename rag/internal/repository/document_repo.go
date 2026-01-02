package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/rag/internal/model"
)

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(ctx context.Context, doc *model.Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *DocumentRepository) CreateBatch(ctx context.Context, docs []model.Document) error {
	return r.db.WithContext(ctx).Create(&docs).Error
}

func (r *DocumentRepository) FindByCollectionID(ctx context.Context, collectionID uuid.UUID, limit, offset int) ([]model.Document, int64, error) {
	var docs []model.Document
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Document{}).
		Where("collection_id = ? AND deleted_at IS NULL", collectionID)

	query.Count(&total)
	err := query.Order("chunk_index ASC").Limit(limit).Offset(offset).Find(&docs).Error
	return docs, total, err
}

func (r *DocumentRepository) CountByCollectionID(ctx context.Context, collectionID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Document{}).
		Where("collection_id = ? AND deleted_at IS NULL", collectionID).
		Count(&count).Error
	return count, err
}

func (r *DocumentRepository) DeleteByFileID(ctx context.Context, fileID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("file_id = ?", fileID).Delete(&model.Document{}).Error
}

func (r *DocumentRepository) DeleteByCollectionID(ctx context.Context, collectionID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("collection_id = ?", collectionID).Delete(&model.Document{}).Error
}

func (r *DocumentRepository) FindByFileID(ctx context.Context, fileID uuid.UUID, limit, offset int) ([]model.Document, int64, error) {
	var docs []model.Document
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Document{}).
		Where("file_id = ? AND deleted_at IS NULL", fileID)

	query.Count(&total)
	err := query.Order("chunk_index ASC").Limit(limit).Offset(offset).Find(&docs).Error
	return docs, total, err
}

// DeleteByQAPairID deletes documents associated with a QA pair
func (r *DocumentRepository) DeleteByQAPairID(ctx context.Context, qaPairID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("metadata->>'qa_pair_id' = ?", qaPairID.String()).
		Delete(&model.Document{}).Error
}

// CreateOrUpdateByQAPairID creates or updates a document for a QA pair
func (r *DocumentRepository) CreateOrUpdateByQAPairID(ctx context.Context, qaPairID uuid.UUID, doc *model.Document) error {
	var existing model.Document
	err := r.db.WithContext(ctx).
		Where("metadata->>'qa_pair_id' = ?", qaPairID.String()).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new document
		return r.db.WithContext(ctx).Create(doc).Error
	} else if err != nil {
		return err
	}

	// Update existing document
	existing.Content = doc.Content
	existing.Embedding = doc.Embedding
	existing.Metadata = doc.Metadata
	return r.db.WithContext(ctx).Save(&existing).Error
}
