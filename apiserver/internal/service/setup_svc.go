package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/repository"
)

// makeSlug creates a URL-friendly slug from a string
func makeSlug(s string) string {
	s = strings.ToLower(s)
	s = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(s, "-")
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "default"
	}
	return s
}

var (
	ErrAlreadyInstalled = errors.New("system is already installed")
	ErrAdminExists      = errors.New("admin already exists")
	ErrAdminRequired    = errors.New("admin account must be created first")
)

type SetupService struct {
	db           *gorm.DB
	setupRepo    *repository.SetupRepository
	platformRepo *repository.PlatformRepository
}

func NewSetupService(db *gorm.DB, setupRepo *repository.SetupRepository, platformRepo *repository.PlatformRepository) *SetupService {
	return &SetupService{
		db:           db,
		setupRepo:    setupRepo,
		platformRepo: platformRepo,
	}
}

// SetupStatusResponse represents the setup status response
type SetupStatusResponse struct {
	IsInstalled      bool       `json:"is_installed"`
	HasAdmin         bool       `json:"has_admin"`
	HasUserStaff     bool       `json:"has_user_staff"`
	HasLLMConfig     bool       `json:"has_llm_config"`
	SkipLLMConfig    bool       `json:"skip_llm_config"`
	SetupCompletedAt *time.Time `json:"setup_completed_at,omitempty"`
}

// GetStatus returns the current setup status
func (s *SetupService) GetStatus(ctx context.Context) (*SetupStatusResponse, error) {
	setup, err := s.setupRepo.GetOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	setup.RecalculateInstallFlags()
	if err := s.setupRepo.Update(ctx, setup); err != nil {
		return nil, err
	}

	hasAdmin, err := s.setupRepo.HasAdmin(ctx)
	if err != nil {
		return nil, err
	}

	hasUserStaff, err := s.setupRepo.HasUserStaff(ctx)
	if err != nil {
		return nil, err
	}

	return &SetupStatusResponse{
		IsInstalled:      setup.IsInstalled,
		HasAdmin:         hasAdmin,
		HasUserStaff:     hasUserStaff,
		HasLLMConfig:     setup.LLMConfigured,
		SkipLLMConfig:    setup.SkipLLMConfig,
		SetupCompletedAt: setup.SetupCompletedAt,
	}, nil
}

// CreateAdminRequest represents the create admin request
type CreateAdminRequest struct {
	Password      string `json:"password" binding:"required,min=6"`
	Nickname      string `json:"nickname"`
	ProjectName   string `json:"project_name"`
	SkipLLMConfig bool   `json:"skip_llm_config"`
}

// CreateAdminResponse represents the create admin response
type CreateAdminResponse struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Nickname    string    `json:"nickname"`
	ProjectID   uuid.UUID `json:"project_id"`
	ProjectName string    `json:"project_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateAdmin creates the first admin account
func (s *SetupService) CreateAdmin(ctx context.Context, req *CreateAdminRequest) (*CreateAdminResponse, error) {
	setup, err := s.setupRepo.GetOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	// Check if admin already exists (idempotent)
	existingAdmin, err := s.setupRepo.GetAdminByUsername(ctx, "admin")
	if err == nil && existingAdmin != nil {
		projectID := uuid.Nil
		if existingAdmin.ProjectID != nil {
			projectID = *existingAdmin.ProjectID
		}
		return &CreateAdminResponse{
			ID:          existingAdmin.ID,
			Username:    "admin",
			Nickname:    existingAdmin.FullName,
			ProjectID:   projectID,
			ProjectName: "Default Project",
			CreatedAt:   existingAdmin.CreatedAt,
		}, nil
	}

	if setup.IsInstalled {
		return nil, ErrAlreadyInstalled
	}

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create default project
	projectName := req.ProjectName
	if projectName == "" {
		projectName = "Default Project"
	}

	project := &model.Project{
		Name:   projectName,
		Slug:   makeSlug(projectName),
		APIKey: generateAPIKey(),
	}
	if err := tx.Create(project).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create admin staff
	nickname := req.Nickname
	if nickname == "" {
		nickname = "Administrator"
	}

	projectID := project.ID
	admin := &model.Staff{
		ProjectID:    &projectID,
		Username:     "admin",
		Email:        "admin@placeholder.local",
		PasswordHash: string(passwordHash),
		FullName:     nickname,
		Role:         "admin",
		IsActive:     true,
		IsSuperAdmin: true,
	}
	if err := tx.Create(admin).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create default website platform
	websiteAPIKey := generatePlatformAPIKey()
	websitePlatform := &model.Platform{
		ProjectID: project.ID,
		Type:      "website",
		APIKey:    &websiteAPIKey,
		Config: map[string]interface{}{
			"position":        "bottom-right",
			"welcome_message": "Hello! How can I help you today?",
			"widget_title":    "TGO AI Chatbot",
		},
		IsActive: true,
	}
	if err := tx.Create(websitePlatform).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update setup flags
	setup.AdminCreated = true
	if req.SkipLLMConfig {
		setup.SkipLLMConfig = true
	}
	setup.RecalculateInstallFlags()
	if err := tx.Save(setup).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &CreateAdminResponse{
		ID:          admin.ID,
		Username:    "admin",
		Nickname:    admin.FullName,
		ProjectID:   project.ID,
		ProjectName: project.Name,
		CreatedAt:   admin.CreatedAt,
	}, nil
}

// SkipLLMConfig skips LLM configuration
func (s *SetupService) SkipLLMConfig(ctx context.Context) error {
	setup, err := s.setupRepo.GetOrCreate(ctx)
	if err != nil {
		return err
	}

	if setup.IsInstalled {
		return ErrAlreadyInstalled
	}

	setup.SkipLLMConfig = true
	setup.RecalculateInstallFlags()
	return s.setupRepo.Update(ctx, setup)
}

// MarkLLMConfigured marks LLM as configured
func (s *SetupService) MarkLLMConfigured(ctx context.Context) error {
	setup, err := s.setupRepo.GetOrCreate(ctx)
	if err != nil {
		return err
	}

	setup.LLMConfigured = true
	setup.RecalculateInstallFlags()
	return s.setupRepo.Update(ctx, setup)
}

// ConfigureLLMRequest represents the configure LLM request
type ConfigureLLMRequest struct {
	Provider        string                 `json:"provider" binding:"required"`
	Name            string                 `json:"name" binding:"required"`
	APIKey          string                 `json:"api_key" binding:"required"`
	APIBaseURL      string                 `json:"api_base_url"`
	DefaultModel    string                 `json:"default_model"`
	AvailableModels []string               `json:"available_models"`
	Config          map[string]interface{} `json:"config"`
	IsActive        bool                   `json:"is_active"`
}

// ConfigureLLMResponse represents the configure LLM response
type ConfigureLLMResponse struct {
	ID           uuid.UUID `json:"id"`
	Provider     string    `json:"provider"`
	Name         string    `json:"name"`
	DefaultModel string    `json:"default_model"`
	IsActive     bool      `json:"is_active"`
	ProjectID    uuid.UUID `json:"project_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// ConfigureLLM configures the LLM provider
func (s *SetupService) ConfigureLLM(ctx context.Context, req *ConfigureLLMRequest) (*ConfigureLLMResponse, error) {
	setup, err := s.setupRepo.GetOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	if setup.IsInstalled {
		return nil, ErrAlreadyInstalled
	}

	if !setup.AdminCreated {
		return nil, ErrAdminRequired
	}

	// Get the first project
	project, err := s.setupRepo.GetFirstProject(ctx)
	if err != nil {
		return nil, err
	}

	// Create AI Provider
	aiProvider := &model.AIProvider{
		ProjectID:       project.ID,
		Provider:        req.Provider,
		Name:            req.Name,
		APIKey:          req.APIKey,
		APIBaseURL:      req.APIBaseURL,
		DefaultModel:    req.DefaultModel,
		AvailableModels: req.AvailableModels,
		Config:          req.Config,
		IsActive:        true,
	}

	if err := s.setupRepo.CreateAIProvider(ctx, aiProvider); err != nil {
		return nil, err
	}

	// Update setup flags
	setup.LLMConfigured = true
	setup.SkipLLMConfig = false
	setup.RecalculateInstallFlags()
	if err := s.setupRepo.Update(ctx, setup); err != nil {
		return nil, err
	}

	return &ConfigureLLMResponse{
		ID:           aiProvider.ID,
		Provider:     aiProvider.Provider,
		Name:         aiProvider.Name,
		DefaultModel: aiProvider.DefaultModel,
		IsActive:     aiProvider.IsActive,
		ProjectID:    aiProvider.ProjectID,
		CreatedAt:    aiProvider.CreatedAt,
	}, nil
}

// StaffItem represents a single staff member to create
type StaffItem struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
	Name        string `json:"name"`
	Nickname    string `json:"nickname"`
	Description string `json:"description"`
}

// BatchCreateStaffRequest represents the batch create staff request
type BatchCreateStaffRequest struct {
	StaffList []StaffItem `json:"staff_list" binding:"required"`
}

// StaffCreatedItem represents a created staff member
type StaffCreatedItem struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Nickname  string    `json:"nickname"`
	CreatedAt time.Time `json:"created_at"`
}

// BatchCreateStaffResponse represents the batch create staff response
type BatchCreateStaffResponse struct {
	CreatedCount     int                `json:"created_count"`
	StaffList        []StaffCreatedItem `json:"staff_list"`
	SkippedUsernames []string           `json:"skipped_usernames"`
}

// BatchCreateStaff creates staff members during setup
func (s *SetupService) BatchCreateStaff(ctx context.Context, req *BatchCreateStaffRequest) (*BatchCreateStaffResponse, error) {
	setup, err := s.setupRepo.GetOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	if setup.IsInstalled {
		return nil, ErrAlreadyInstalled
	}

	if !setup.AdminCreated {
		return nil, ErrAdminRequired
	}

	// Get the first project
	project, err := s.setupRepo.GetFirstProject(ctx)
	if err != nil {
		return nil, err
	}

	createdStaff := []StaffCreatedItem{}
	skippedUsernames := []string{}

	for _, item := range req.StaffList {
		// Check if username already exists
		_, err := s.setupRepo.GetStaffByUsername(ctx, item.Username)
		if err == nil {
			skippedUsernames = append(skippedUsernames, item.Username)
			continue
		}

		// Hash password
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(item.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}

		nickname := item.Nickname
		if nickname == "" {
			nickname = item.Name
		}
		if nickname == "" {
			nickname = item.Username
		}

		projectID := project.ID
		staff := &model.Staff{
			ProjectID:    &projectID,
			Username:     item.Username,
			Email:        item.Username + "@placeholder.local",
			PasswordHash: string(passwordHash),
			FullName:     item.Name,
			Role:         "user",
			IsActive:     true,
		}

		if err := s.setupRepo.CreateStaff(ctx, staff); err != nil {
			return nil, err
		}

		createdStaff = append(createdStaff, StaffCreatedItem{
			ID:        staff.ID,
			Username:  staff.Username,
			Name:      staff.FullName,
			Nickname:  nickname,
			CreatedAt: staff.CreatedAt,
		})
	}

	return &BatchCreateStaffResponse{
		CreatedCount:     len(createdStaff),
		StaffList:        createdStaff,
		SkippedUsernames: skippedUsernames,
	}, nil
}
