package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/pkg/aicenter"
)

type MCPToolsHandler struct {
	client *aicenter.Client
}

func NewMCPToolsHandler(client *aicenter.Client) *MCPToolsHandler {
	return &MCPToolsHandler{client: client}
}

// List lists project tools
func (h *MCPToolsHandler) List(c *gin.Context) {
	projectID, _ := c.Get("project_id")

	params := map[string]string{}
	if v := c.Query("source_type"); v != "" {
		params["source_type"] = v
	}
	if v := c.Query("status"); v != "" {
		params["status"] = v
	}
	if v := c.Query("is_enabled"); v != "" {
		params["is_enabled"] = v
	}
	if v := c.Query("search"); v != "" {
		params["search"] = v
	}
	if v := c.Query("page"); v != "" {
		params["page"] = v
	}
	if v := c.Query("per_page"); v != "" {
		params["per_page"] = v
	}

	result, err := h.client.ListProjectTools(c.Request.Context(), projectID.(uuid.UUID).String(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetStats gets project tool statistics
func (h *MCPToolsHandler) GetStats(c *gin.Context) {
	projectID, _ := c.Get("project_id")

	result, err := h.client.GetProjectToolStats(c.Request.Context(), projectID.(uuid.UUID).String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Get gets a specific project tool
func (h *MCPToolsHandler) Get(c *gin.Context) {
	projectID, _ := c.Get("project_id")
	toolID := c.Param("id")

	result, err := h.client.GetProjectTool(c.Request.Context(), projectID.(uuid.UUID).String(), toolID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tool not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Update updates a project tool
func (h *MCPToolsHandler) Update(c *gin.Context) {
	projectID, _ := c.Get("project_id")
	toolID := c.Param("id")

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.client.UpdateProjectTool(c.Request.Context(), projectID.(uuid.UUID).String(), toolID, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Uninstall uninstalls a tool from the project
func (h *MCPToolsHandler) Uninstall(c *gin.Context) {
	projectID, _ := c.Get("project_id")
	toolID := c.Param("id")

	err := h.client.UninstallTool(c.Request.Context(), projectID.(uuid.UUID).String(), toolID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// Install installs a tool to the project
func (h *MCPToolsHandler) Install(c *gin.Context) {
	projectID, _ := c.Get("project_id")

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.client.InstallTool(c.Request.Context(), projectID.(uuid.UUID).String(), data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// BulkInstall bulk installs tools to the project
func (h *MCPToolsHandler) BulkInstall(c *gin.Context) {
	projectID, _ := c.Get("project_id")

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.client.BulkInstallTools(c.Request.Context(), projectID.(uuid.UUID).String(), data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}
