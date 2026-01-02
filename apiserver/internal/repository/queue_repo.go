package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
)

type QueueRepository struct {
	DB *gorm.DB
}

func NewQueueRepository(db *gorm.DB) *QueueRepository {
	return &QueueRepository{DB: db}
}

func (r *QueueRepository) Create(ctx context.Context, item *model.VisitorWaitingQueue) error {
	return r.DB.WithContext(ctx).Create(item).Error
}

func (r *QueueRepository) FindWaiting(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.VisitorWaitingQueue, int64, error) {
	var items []model.VisitorWaitingQueue
	var total int64

	query := r.DB.WithContext(ctx).Model(&model.VisitorWaitingQueue{}).
		Where("project_id = ? AND status = ? AND deleted_at IS NULL", projectID, model.QueueStatusWaiting).
		Preload("Visitor")

	query.Count(&total)
	err := query.Order("priority DESC, wait_started_at ASC").Limit(limit).Offset(offset).Find(&items).Error
	return items, total, err
}

func (r *QueueRepository) FindByID(ctx context.Context, projectID, id uuid.UUID) (*model.VisitorWaitingQueue, error) {
	var item model.VisitorWaitingQueue
	err := r.DB.WithContext(ctx).
		Where("id = ? AND project_id = ? AND deleted_at IS NULL", id, projectID).
		Preload("Visitor").
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *QueueRepository) Assign(ctx context.Context, projectID, id, staffID uuid.UUID) error {
	now := time.Now()
	return r.DB.WithContext(ctx).Model(&model.VisitorWaitingQueue{}).
		Where("id = ? AND project_id = ?", id, projectID).
		Updates(map[string]interface{}{
			"status":      model.QueueStatusAssigned,
			"assigned_to": staffID,
			"assigned_at": now,
		}).Error
}

func (r *QueueRepository) UpdateStatus(ctx context.Context, projectID, id uuid.UUID, status model.QueueStatus) error {
	return r.DB.WithContext(ctx).Model(&model.VisitorWaitingQueue{}).
		Where("id = ? AND project_id = ?", id, projectID).
		Update("status", status).Error
}

func (r *QueueRepository) GetPosition(ctx context.Context, projectID, id uuid.UUID) (int64, error) {
	var item model.VisitorWaitingQueue
	err := r.DB.WithContext(ctx).
		Where("id = ? AND project_id = ?", id, projectID).
		First(&item).Error
	if err != nil {
		return 0, err
	}

	var position int64
	r.DB.WithContext(ctx).Model(&model.VisitorWaitingQueue{}).
		Where("project_id = ? AND status = ? AND wait_started_at < ? AND deleted_at IS NULL",
			projectID, model.QueueStatusWaiting, item.WaitStartedAt).
		Count(&position)

	return position + 1, nil
}

func (r *QueueRepository) GetCount(ctx context.Context, projectID uuid.UUID) (waiting, assigned int64, err error) {
	r.DB.WithContext(ctx).Model(&model.VisitorWaitingQueue{}).
		Where("project_id = ? AND status = ? AND deleted_at IS NULL", projectID, model.QueueStatusWaiting).
		Count(&waiting)

	r.DB.WithContext(ctx).Model(&model.VisitorWaitingQueue{}).
		Where("project_id = ? AND status = ? AND deleted_at IS NULL", projectID, model.QueueStatusAssigned).
		Count(&assigned)

	return waiting, assigned, nil
}

func (r *QueueRepository) CountWaiting(ctx context.Context, projectID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(&model.VisitorWaitingQueue{}).
		Where("project_id = ? AND status = ? AND visitor_id IS NOT NULL AND deleted_at IS NULL", projectID, model.QueueStatusWaiting).
		Count(&count).Error
	return count, err
}

func (r *QueueRepository) FindWaitingPaginated(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.VisitorWaitingQueue, error) {
	var items []model.VisitorWaitingQueue
	err := r.DB.WithContext(ctx).
		Where("project_id = ? AND status = ? AND visitor_id IS NOT NULL AND deleted_at IS NULL", projectID, model.QueueStatusWaiting).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&items).Error
	return items, err
}

// FindByVisitorID finds queue entry by visitor ID
func (r *QueueRepository) FindByVisitorID(ctx context.Context, projectID, visitorID uuid.UUID) (*model.VisitorWaitingQueue, error) {
	var item model.VisitorWaitingQueue
	err := r.DB.WithContext(ctx).
		Where("project_id = ? AND visitor_id = ? AND status = ? AND deleted_at IS NULL",
			projectID, visitorID, model.QueueStatusWaiting).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Delete soft deletes a queue entry
func (r *QueueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.DB.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.VisitorWaitingQueue{}).Error
}
