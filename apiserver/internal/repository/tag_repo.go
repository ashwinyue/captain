package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
)

type TagRepository struct {
	BaseRepository[model.Tag]
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{BaseRepository: BaseRepository[model.Tag]{DB: db}}
}

func (r *TagRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID, category string, limit, offset int) ([]model.Tag, int64, error) {
	var tags []model.Tag
	var total int64

	query := r.DB.WithContext(ctx).Model(&model.Tag{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	query.Count(&total)
	err := query.Order("sort_order ASC, created_at DESC").Limit(limit).Offset(offset).Find(&tags).Error
	return tags, total, err
}

func (r *TagRepository) FindByIDAndProject(ctx context.Context, projectID, id uuid.UUID) (*model.Tag, error) {
	var tag model.Tag
	err := r.DB.WithContext(ctx).
		Where("id = ? AND project_id = ? AND deleted_at IS NULL", id, projectID).
		First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

type VisitorTagRepository struct {
	DB *gorm.DB
}

func NewVisitorTagRepository(db *gorm.DB) *VisitorTagRepository {
	return &VisitorTagRepository{DB: db}
}

func (r *VisitorTagRepository) AddTagToVisitor(ctx context.Context, projectID, visitorID, tagID uuid.UUID) error {
	vt := &model.VisitorTag{
		ProjectID: projectID,
		VisitorID: visitorID,
		TagID:     tagID,
	}
	return r.DB.WithContext(ctx).Create(vt).Error
}

func (r *VisitorTagRepository) RemoveTagFromVisitor(ctx context.Context, projectID, visitorID, tagID uuid.UUID) error {
	return r.DB.WithContext(ctx).
		Where("project_id = ? AND visitor_id = ? AND tag_id = ?", projectID, visitorID, tagID).
		Delete(&model.VisitorTag{}).Error
}

func (r *VisitorTagRepository) GetVisitorTags(ctx context.Context, projectID, visitorID uuid.UUID) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.DB.WithContext(ctx).
		Joins("JOIN visitor_tags ON tags.id = visitor_tags.tag_id").
		Where("visitor_tags.project_id = ? AND visitor_tags.visitor_id = ? AND tags.deleted_at IS NULL",
			projectID, visitorID).
		Find(&tags).Error
	return tags, err
}
