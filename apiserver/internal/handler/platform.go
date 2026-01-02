package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/service"
)

type PlatformHandler struct {
	svc *service.PlatformService
}

func NewPlatformHandler(svc *service.PlatformService) *PlatformHandler {
	return &PlatformHandler{svc: svc}
}

// ListTypes lists all platform type definitions
func (h *PlatformHandler) ListTypes(c *gin.Context) {
	types, err := h.svc.ListTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, types)
}

// List lists platforms for a project
func (h *PlatformHandler) List(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))

	// Query params
	platformType := c.Query("type")
	var isActive *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		active := activeStr == "true"
		isActive = &active
	}

	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	platforms, total, err := h.svc.List(c.Request.Context(), projectID, platformType, isActive, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": platforms,
		"pagination": gin.H{
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"has_next": int64(offset+limit) < total,
			"has_prev": offset > 0,
		},
	})
}

// GetInfo gets platform info by API key (visitor-facing)
func (h *PlatformHandler) GetInfo(c *gin.Context) {
	apiKey := c.Query("platform_api_key")
	if apiKey == "" {
		apiKey = c.GetHeader("X-Platform-API-Key")
	}

	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Missing platform_api_key"})
		return
	}

	platform, err := h.svc.GetByAPIKey(c.Request.Context(), apiKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid platform_api_key"})
		return
	}

	if platform.IsDeleted() {
		c.JSON(http.StatusForbidden, gin.H{"detail": "Platform is deleted"})
		return
	}

	if !platform.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"detail": "Platform is disabled"})
		return
	}

	c.JSON(http.StatusOK, platform)
}

// Get gets a platform by ID
func (h *PlatformHandler) Get(c *gin.Context) {
	projectID, _ := c.Get("project_id")
	platformID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid platform ID"})
		return
	}

	platform, err := h.svc.GetByID(c.Request.Context(), projectID.(uuid.UUID), platformID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Platform not found"})
		return
	}

	c.JSON(http.StatusOK, platform)
}

// Create creates a new platform
func (h *PlatformHandler) Create(c *gin.Context) {
	projectID, _ := c.Get("project_id")

	var req service.CreatePlatformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	platform, err := h.svc.Create(c.Request.Context(), projectID.(uuid.UUID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, platform)
}

// Update updates a platform
func (h *PlatformHandler) Update(c *gin.Context) {
	projectID, _ := c.Get("project_id")
	platformID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid platform ID"})
		return
	}

	var req service.UpdatePlatformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	platform, err := h.svc.Update(c.Request.Context(), projectID.(uuid.UUID), platformID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, platform)
}

// Delete deletes a platform
func (h *PlatformHandler) Delete(c *gin.Context) {
	projectID, _ := c.Get("project_id")
	platformID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid platform ID"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), projectID.(uuid.UUID), platformID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// RegenerateAPIKey regenerates the API key for a platform
func (h *PlatformHandler) RegenerateAPIKey(c *gin.Context) {
	projectID, _ := c.Get("project_id")
	platformID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid platform ID"})
		return
	}

	platform, err := h.svc.RegenerateAPIKey(c.Request.Context(), projectID.(uuid.UUID), platformID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"api_key": platform.APIKey})
}
