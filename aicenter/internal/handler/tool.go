package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/aicenter/internal/model"
	"github.com/tgo/captain/aicenter/internal/pkg/response"
	"github.com/tgo/captain/aicenter/internal/service"
)

type ToolHandler struct {
	svc *service.ToolService
}

func NewToolHandler(svc *service.ToolService) *ToolHandler {
	return &ToolHandler{svc: svc}
}

func (h *ToolHandler) List(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	includeDeleted, _ := strconv.ParseBool(c.DefaultQuery("include_deleted", "false"))

	var toolType *model.ToolType
	if tt := c.Query("tool_type"); tt != "" {
		t := model.ToolType(tt)
		toolType = &t
	}

	tools, total, err := h.svc.List(c.Request.Context(), projectID, toolType, includeDeleted, limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.List(c, tools, total, limit, offset)
}

func (h *ToolHandler) Create(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	var req model.Tool
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

func (h *ToolHandler) Get(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	toolID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tool id")
		return
	}

	tool, err := h.svc.GetByID(c.Request.Context(), projectID, toolID)
	if err != nil {
		response.NotFound(c, "TOOL")
		return
	}

	response.Success(c, tool)
}

func (h *ToolHandler) Update(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	toolID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tool id")
		return
	}

	tool, err := h.svc.GetByID(c.Request.Context(), projectID, toolID)
	if err != nil {
		response.NotFound(c, "TOOL")
		return
	}

	if err := c.ShouldBindJSON(tool); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.Update(c.Request.Context(), tool); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, tool)
}

func (h *ToolHandler) Delete(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	toolID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tool id")
		return
	}

	// Get tool first for response
	tool, err := h.svc.GetByID(c.Request.Context(), projectID, toolID)
	if err != nil {
		response.NotFound(c, "TOOL")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), projectID, toolID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, tool)
}
