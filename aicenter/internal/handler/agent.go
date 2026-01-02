package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/aicenter/internal/model"
	"github.com/tgo/captain/aicenter/internal/pkg/response"
	"github.com/tgo/captain/aicenter/internal/service"
)

type AgentHandler struct {
	svc *service.AgentService
}

func NewAgentHandler(svc *service.AgentService) *AgentHandler {
	return &AgentHandler{svc: svc}
}

func (h *AgentHandler) List(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var teamID *uuid.UUID
	if tid := c.Query("team_id"); tid != "" {
		if parsed, err := uuid.Parse(tid); err == nil {
			teamID = &parsed
		}
	}

	agents, total, err := h.svc.List(c.Request.Context(), projectID, teamID, limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.List(c, agents, total, limit, offset)
}

func (h *AgentHandler) Create(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	var req model.Agent
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req.ProjectID = projectID
	if err := h.svc.Create(c.Request.Context(), &req); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, req)
}

func (h *AgentHandler) Get(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	agentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid agent id")
		return
	}

	agent, err := h.svc.GetByID(c.Request.Context(), projectID, agentID)
	if err != nil {
		response.NotFound(c, "AGENT")
		return
	}

	response.Success(c, agent)
}

// AgentUpdateRequest handles collection_ids as list of strings
type AgentUpdateRequest struct {
	Name          string                 `json:"name,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Instruction   string                 `json:"instruction,omitempty"`
	Model         string                 `json:"model,omitempty"`
	IsDefault     *bool                  `json:"is_default,omitempty"`
	IsEnabled     *bool                  `json:"is_enabled,omitempty"`
	Config        map[string]interface{} `json:"config,omitempty"`
	TeamID        *uuid.UUID             `json:"team_id,omitempty"`
	LLMProviderID *uuid.UUID             `json:"llm_provider_id,omitempty"`
	Collections   []string               `json:"collections,omitempty"` // Collection IDs as strings
	Tools         []model.AgentTool      `json:"tools,omitempty"`
}

func (h *AgentHandler) Update(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	agentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid agent id")
		return
	}

	agent, err := h.svc.GetByID(c.Request.Context(), projectID, agentID)
	if err != nil {
		response.NotFound(c, "AGENT")
		return
	}

	var req AgentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Update agent fields
	if req.Name != "" {
		agent.Name = req.Name
	}
	if req.Description != "" {
		agent.Description = req.Description
	}
	if req.Instruction != "" {
		agent.Instruction = req.Instruction
	}
	if req.Model != "" {
		agent.Model = req.Model
	}
	if req.IsDefault != nil {
		agent.IsDefault = *req.IsDefault
	}
	if req.IsEnabled != nil {
		agent.IsEnabled = *req.IsEnabled
	}
	if req.Config != nil {
		agent.Config = req.Config
	}
	if req.TeamID != nil {
		agent.TeamID = req.TeamID
	}
	if req.LLMProviderID != nil {
		agent.LLMProviderID = req.LLMProviderID
	}

	// Convert collection IDs to AgentCollection objects
	if req.Collections != nil {
		agent.Collections = make([]model.AgentCollection, len(req.Collections))
		for i, collID := range req.Collections {
			agent.Collections[i] = model.AgentCollection{
				AgentID:      agentID,
				CollectionID: collID,
				IsEnabled:    true,
			}
		}
	}

	// Update tools if provided
	if req.Tools != nil {
		agent.Tools = req.Tools
	}

	if err := h.svc.Update(c.Request.Context(), agent); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, agent)
}

func (h *AgentHandler) Delete(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	agentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid agent id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), projectID, agentID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.NoContent(c)
}

func (h *AgentHandler) Exists(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	exists, count, err := h.svc.Exists(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"exists": exists, "count": count})
}

func (h *AgentHandler) SetToolEnabled(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	agentID, _ := uuid.Parse(c.Param("id"))
	toolID, _ := uuid.Parse(c.Param("tool_id"))

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.SetToolEnabled(c.Request.Context(), projectID, agentID, toolID, req.Enabled); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.NoContent(c)
}

func (h *AgentHandler) SetCollectionEnabled(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	agentID, _ := uuid.Parse(c.Param("id"))
	collectionID := c.Param("collection_id")

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.SetCollectionEnabled(c.Request.Context(), projectID, agentID, collectionID, req.Enabled); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.NoContent(c)
}
