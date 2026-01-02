package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/aicenter/internal/model"
	"github.com/tgo/captain/aicenter/internal/pkg/response"
	"github.com/tgo/captain/aicenter/internal/service"
)

type ProjectAIConfigHandler struct {
	svc *service.ProjectAIConfigService
}

func NewProjectAIConfigHandler(svc *service.ProjectAIConfigService) *ProjectAIConfigHandler {
	return &ProjectAIConfigHandler{svc: svc}
}

// ProjectAIConfigUpsert request body
type ProjectAIConfigUpsert struct {
	ProjectID                  uuid.UUID     `json:"project_id" binding:"required"`
	DefaultChatProviderID      *uuid.UUID    `json:"default_chat_provider_id"`
	DefaultChatModel           string        `json:"default_chat_model"`
	DefaultEmbeddingProviderID *uuid.UUID    `json:"default_embedding_provider_id"`
	DefaultEmbeddingModel      string        `json:"default_embedding_model"`
	DefaultTeamID              *uuid.UUID    `json:"default_team_id"`
	Config                     model.JSONMap `json:"config"`
}

// ProjectAIConfigSyncRequest for bulk sync
type ProjectAIConfigSyncRequest struct {
	Configs []ProjectAIConfigUpsert `json:"configs" binding:"required"`
}

// Sync handles bulk upsert of project AI configs
func (h *ProjectAIConfigHandler) Sync(c *gin.Context) {
	var req ProjectAIConfigSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	configs := make([]*model.ProjectAIConfig, 0, len(req.Configs))
	for _, cfg := range req.Configs {
		configs = append(configs, &model.ProjectAIConfig{
			ProjectID:                  cfg.ProjectID,
			DefaultChatProviderID:      cfg.DefaultChatProviderID,
			DefaultChatModel:           cfg.DefaultChatModel,
			DefaultEmbeddingProviderID: cfg.DefaultEmbeddingProviderID,
			DefaultEmbeddingModel:      cfg.DefaultEmbeddingModel,
			DefaultTeamID:              cfg.DefaultTeamID,
			Config:                     cfg.Config,
		})
	}

	if err := h.svc.SyncConfigs(c.Request.Context(), configs); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// Return the synced configs
	c.JSON(http.StatusOK, gin.H{
		"data": configs,
	})
}

// Upsert handles single project AI config upsert
func (h *ProjectAIConfigHandler) Upsert(c *gin.Context) {
	var req ProjectAIConfigUpsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Check if exists
	existing, _ := h.svc.GetByProjectID(c.Request.Context(), req.ProjectID)

	config := &model.ProjectAIConfig{
		ProjectID:                  req.ProjectID,
		DefaultChatProviderID:      req.DefaultChatProviderID,
		DefaultChatModel:           req.DefaultChatModel,
		DefaultEmbeddingProviderID: req.DefaultEmbeddingProviderID,
		DefaultEmbeddingModel:      req.DefaultEmbeddingModel,
		DefaultTeamID:              req.DefaultTeamID,
		Config:                     req.Config,
	}

	if err := h.svc.Upsert(c.Request.Context(), config); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// Return 201 if created, 200 if updated
	status := http.StatusOK
	if existing == nil {
		status = http.StatusCreated
	}

	c.JSON(status, config)
}

// Get retrieves project AI config by project ID
func (h *ProjectAIConfigHandler) Get(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		projectIDStr = c.GetString("project_id")
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	config, err := h.svc.GetByProjectID(c.Request.Context(), projectID)
	if err != nil {
		response.NotFound(c, "PROJECT_AI_CONFIG")
		return
	}

	response.Success(c, config)
}
