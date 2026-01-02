package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
)

type StaffRepository struct {
	BaseRepository[model.Staff]
}

func NewStaffRepository(db *gorm.DB) *StaffRepository {
	return &StaffRepository{BaseRepository: BaseRepository[model.Staff]{DB: db}}
}

func (r *StaffRepository) FindByUsername(ctx context.Context, username string) (*model.Staff, error) {
	var staff model.Staff
	err := r.DB.WithContext(ctx).
		Where("username = ? AND is_active = true AND deleted_at IS NULL", username).
		First(&staff).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

func (r *StaffRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.Staff, int64, error) {
	var staffs []model.Staff
	var total int64

	query := r.DB.WithContext(ctx).Model(&model.Staff{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID)

	query.Count(&total)
	err := query.Limit(limit).Offset(offset).Find(&staffs).Error
	return staffs, total, err
}

func (r *StaffRepository) FindAllWithProject(ctx context.Context, projectID *uuid.UUID, limit, offset int) ([]model.Staff, int64, error) {
	var staffs []model.Staff
	var total int64

	query := r.DB.WithContext(ctx).Model(&model.Staff{}).Where("deleted_at IS NULL")
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	query.Count(&total)
	err := query.Limit(limit).Offset(offset).Find(&staffs).Error
	return staffs, total, err
}

// FindAvailableStaff finds staff who are active and available for assignment
func (r *StaffRepository) FindAvailableStaff(ctx context.Context, projectID uuid.UUID) ([]model.Staff, error) {
	var staffs []model.Staff
	err := r.DB.WithContext(ctx).
		Where("project_id = ? AND is_active = true AND service_paused = false AND deleted_at IS NULL", projectID).
		Find(&staffs).Error
	return staffs, err
}

// FindByID finds staff by ID
func (r *StaffRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Staff, error) {
	var staff model.Staff
	err := r.DB.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&staff).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}
