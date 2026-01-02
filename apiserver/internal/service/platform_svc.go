package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/repository"
)

type PlatformService struct {
	repo *repository.PlatformRepository
}

func NewPlatformService(repo *repository.PlatformRepository) *PlatformService {
	return &PlatformService{repo: repo}
}

// ListTypes returns all platform type definitions
func (s *PlatformService) ListTypes(ctx context.Context) ([]model.PlatformTypeDefinition, error) {
	return s.repo.ListTypes(ctx)
}

// List returns paginated platforms
func (s *PlatformService) List(ctx context.Context, projectID uuid.UUID, platformType string, isActive *bool, limit, offset int) ([]model.Platform, int64, error) {
	return s.repo.List(ctx, projectID, platformType, isActive, limit, offset)
}

// GetByID returns a platform by ID
func (s *PlatformService) GetByID(ctx context.Context, projectID, platformID uuid.UUID) (*model.Platform, error) {
	return s.repo.GetByID(ctx, projectID, platformID)
}

// GetByAPIKey returns a platform by API key (for visitor-facing info endpoint)
func (s *PlatformService) GetByAPIKey(ctx context.Context, apiKey string) (*model.Platform, error) {
	return s.repo.GetByAPIKey(ctx, apiKey)
}

// CreatePlatformRequest represents a platform creation request
type CreatePlatformRequest struct {
	Name                string                 `json:"name,omitempty"`
	Type                string                 `json:"type"`
	Config              map[string]interface{} `json:"config,omitempty"`
	IsActive            bool                   `json:"is_active"`
	AgentIDs            []uuid.UUID            `json:"agent_ids,omitempty"`
	AIMode              string                 `json:"ai_mode,omitempty"`
	FallbackToAITimeout *int                   `json:"fallback_to_ai_timeout,omitempty"`
}

// Create creates a new platform
func (s *PlatformService) Create(ctx context.Context, projectID uuid.UUID, req *CreatePlatformRequest) (*model.Platform, error) {
	apiKey := generatePlatformAPIKey()

	var name *string
	if req.Name != "" {
		name = &req.Name
	}

	var aiMode *string
	if req.AIMode != "" {
		aiMode = &req.AIMode
	}

	platform := &model.Platform{
		ProjectID:           projectID,
		Name:                name,
		Type:                req.Type,
		APIKey:              &apiKey,
		Config:              req.Config,
		IsActive:            req.IsActive,
		AgentIDs:            req.AgentIDs,
		AIMode:              aiMode,
		FallbackToAITimeout: req.FallbackToAITimeout,
		SyncStatus:          string(model.PlatformSyncPending),
	}

	if err := s.repo.Create(ctx, platform); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, projectID, platform.ID)
}

// UpdatePlatformRequest represents a platform update request
type UpdatePlatformRequest struct {
	Name                *string                `json:"name,omitempty"`
	Config              map[string]interface{} `json:"config,omitempty"`
	IsActive            *bool                  `json:"is_active,omitempty"`
	AgentIDs            []uuid.UUID            `json:"agent_ids,omitempty"`
	AIMode              *string                `json:"ai_mode,omitempty"`
	FallbackToAITimeout *int                   `json:"fallback_to_ai_timeout,omitempty"`
}

// Update updates a platform
func (s *PlatformService) Update(ctx context.Context, projectID, platformID uuid.UUID, req *UpdatePlatformRequest) (*model.Platform, error) {
	platform, err := s.repo.GetByID(ctx, projectID, platformID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		platform.Name = req.Name
	}
	if req.Config != nil {
		platform.Config = req.Config
	}
	if req.IsActive != nil {
		platform.IsActive = *req.IsActive
	}
	if req.AgentIDs != nil {
		platform.AgentIDs = req.AgentIDs
	}
	if req.AIMode != nil {
		platform.AIMode = req.AIMode
	}
	if req.FallbackToAITimeout != nil {
		platform.FallbackToAITimeout = req.FallbackToAITimeout
	}

	// Mark as pending sync after update
	platform.SyncStatus = string(model.PlatformSyncPending)

	if err := s.repo.Update(ctx, platform); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, projectID, platformID)
}

// Delete soft deletes a platform
func (s *PlatformService) Delete(ctx context.Context, projectID, platformID uuid.UUID) error {
	return s.repo.Delete(ctx, projectID, platformID)
}

// RegenerateAPIKey regenerates the API key for a platform
func (s *PlatformService) RegenerateAPIKey(ctx context.Context, projectID, platformID uuid.UUID) (*model.Platform, error) {
	platform, err := s.repo.GetByID(ctx, projectID, platformID)
	if err != nil {
		return nil, err
	}

	apiKey := generatePlatformAPIKey()
	platform.APIKey = &apiKey

	if err := s.repo.Update(ctx, platform); err != nil {
		return nil, err
	}

	return platform, nil
}

// generatePlatformAPIKey generates a random platform API key
func generatePlatformAPIKey() string {
	return generateAPIKey() // reuse existing function with "pk_" prefix handled below
}
