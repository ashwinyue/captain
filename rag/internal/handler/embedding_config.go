package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/rag/internal/service"
)

var _ = service.BatchSyncConfigRequest{} // ensure import is used

type EmbeddingConfigHandler struct {
	svc *service.EmbeddingConfigService
}

func NewEmbeddingConfigHandler(svc *service.EmbeddingConfigService) *EmbeddingConfigHandler {
	return &EmbeddingConfigHandler{svc: svc}
}

func (h *EmbeddingConfigHandler) List(c *gin.Context) {
	configs, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": configs})
}

func (h *EmbeddingConfigHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	config, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetByProjectID returns the active embedding config for a project
func (h *EmbeddingConfigHandler) GetByProjectID(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	config, err := h.svc.GetByProjectID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Active embedding configuration not found for project"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// BatchSync upserts multiple embedding configs
func (h *EmbeddingConfigHandler) BatchSync(c *gin.Context) {
	var req struct {
		Configs []service.BatchSyncConfigRequest `json:"configs"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.svc.BatchSync(c.Request.Context(), req.Configs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
