package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
)

type SessionRepository struct {
	BaseRepository[model.Session]
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{BaseRepository: BaseRepository[model.Session]{DB: db}}
}

func (r *SessionRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID, status *model.SessionStatus, limit, offset int) ([]model.Session, int64, error) {
	var sessions []model.Session
	var total int64

	query := r.DB.WithContext(ctx).Model(&model.Session{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	query.Count(&total)
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&sessions).Error
	return sessions, total, err
}

func (r *SessionRepository) FindByIDAndProject(ctx context.Context, projectID, id uuid.UUID) (*model.Session, error) {
	var session model.Session
	err := r.DB.WithContext(ctx).
		Where("id = ? AND project_id = ? AND deleted_at IS NULL", id, projectID).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) Close(ctx context.Context, projectID, id uuid.UUID) error {
	now := time.Now()
	return r.DB.WithContext(ctx).Model(&model.Session{}).
		Where("id = ? AND project_id = ?", id, projectID).
		Updates(map[string]interface{}{
			"status":   model.SessionStatusClosed,
			"ended_at": now,
		}).Error
}

func (r *SessionRepository) Transfer(ctx context.Context, projectID, id, toStaffID uuid.UUID) error {
	return r.DB.WithContext(ctx).Model(&model.Session{}).
		Where("id = ? AND project_id = ?", id, projectID).
		Update("staff_id", toStaffID).Error
}
