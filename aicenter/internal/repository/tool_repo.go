package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/aicenter/internal/model"
)

type ToolRepository struct {
	db *gorm.DB
}

func NewToolRepository(db *gorm.DB) *ToolRepository {
	return &ToolRepository{db: db}
}

type ToolListOptions struct {
	ToolType       *model.ToolType
	IncludeDeleted bool
	Limit          int
	Offset         int
}

func (r *ToolRepository) List(ctx context.Context, projectID uuid.UUID, opts *ToolListOptions) ([]model.Tool, int64, error) {
	var tools []model.Tool
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Tool{}).Where("project_id = ?", projectID)

	if opts != nil {
		if !opts.IncludeDeleted {
			query = query.Where("deleted_at IS NULL")
		}
		if opts.ToolType != nil {
			query = query.Where("tool_type = ?", *opts.ToolType)
		}
	} else {
		query = query.Where("deleted_at IS NULL")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil && opts.Limit > 0 {
		query = query.Limit(opts.Limit).Offset(opts.Offset)
	}

	if err := query.Find(&tools).Error; err != nil {
		return nil, 0, err
	}

	return tools, total, nil
}

func (r *ToolRepository) GetByID(ctx context.Context, projectID, toolID uuid.UUID) (*model.Tool, error) {
	var tool model.Tool
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND deleted_at IS NULL", toolID, projectID).
		First(&tool).Error
	if err != nil {
		return nil, err
	}
	return &tool, nil
}

func (r *ToolRepository) Create(ctx context.Context, tool *model.Tool) error {
	return r.db.WithContext(ctx).Create(tool).Error
}

func (r *ToolRepository) Update(ctx context.Context, tool *model.Tool) error {
	return r.db.WithContext(ctx).Save(tool).Error
}

func (r *ToolRepository) Delete(ctx context.Context, projectID, toolID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ?", toolID, projectID).
		Delete(&model.Tool{}).Error
}
