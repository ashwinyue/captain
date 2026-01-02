package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
)

type SetupRepository struct {
	db *gorm.DB
}

func NewSetupRepository(db *gorm.DB) *SetupRepository {
	return &SetupRepository{db: db}
}

// GetOrCreate returns the singleton SystemSetup record, creating if not exists
func (r *SetupRepository) GetOrCreate(ctx context.Context) (*model.SystemSetup, error) {
	var setup model.SystemSetup
	err := r.db.WithContext(ctx).Order("created_at ASC").First(&setup).Error
	if err == gorm.ErrRecordNotFound {
		now := time.Now()
		setup = model.SystemSetup{
			IsInstalled:   false,
			AdminCreated:  false,
			LLMConfigured: false,
			SkipLLMConfig: false,
			SetupVersion:  "v1",
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := r.db.WithContext(ctx).Create(&setup).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &setup, nil
}

// Update updates the setup record
func (r *SetupRepository) Update(ctx context.Context, setup *model.SystemSetup) error {
	return r.db.WithContext(ctx).Save(setup).Error
}

// HasAdmin checks if admin user exists
func (r *SetupRepository) HasAdmin(ctx context.Context) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Staff{}).
		Where("username = ? AND deleted_at IS NULL", "admin").
		Count(&count).Error
	return count > 0, err
}

// HasUserStaff checks if any non-admin staff exists
func (r *SetupRepository) HasUserStaff(ctx context.Context) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Staff{}).
		Where("role = ? AND deleted_at IS NULL", "user").
		Count(&count).Error
	return count > 0, err
}

// GetAdminByUsername returns the admin staff by username
func (r *SetupRepository) GetAdminByUsername(ctx context.Context, username string) (*model.Staff, error) {
	var staff model.Staff
	err := r.db.WithContext(ctx).
		Preload("Project").
		Where("username = ? AND deleted_at IS NULL", username).
		First(&staff).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

// GetFirstProject returns the first project
func (r *SetupRepository) GetFirstProject(ctx context.Context) (*model.Project, error) {
	var project model.Project
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetStaffByUsername returns staff by username
func (r *SetupRepository) GetStaffByUsername(ctx context.Context, username string) (*model.Staff, error) {
	var staff model.Staff
	err := r.db.WithContext(ctx).
		Where("username = ? AND deleted_at IS NULL", username).
		First(&staff).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

// CreateProject creates a new project
func (r *SetupRepository) CreateProject(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

// CreateStaff creates a new staff member
func (r *SetupRepository) CreateStaff(ctx context.Context, staff *model.Staff) error {
	return r.db.WithContext(ctx).Create(staff).Error
}

// CreatePlatform creates a new platform
func (r *SetupRepository) CreatePlatform(ctx context.Context, platform *model.Platform) error {
	return r.db.WithContext(ctx).Create(platform).Error
}

// CreateAIProvider creates a new AI provider
func (r *SetupRepository) CreateAIProvider(ctx context.Context, provider *model.AIProvider) error {
	return r.db.WithContext(ctx).Create(provider).Error
}
