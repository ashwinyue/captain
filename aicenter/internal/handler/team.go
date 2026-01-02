package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/aicenter/internal/model"
	"github.com/tgo/captain/aicenter/internal/pkg/response"
	"github.com/tgo/captain/aicenter/internal/service"
)

type TeamHandler struct {
	svc *service.TeamService
}

func NewTeamHandler(svc *service.TeamService) *TeamHandler {
	return &TeamHandler{svc: svc}
}

func (h *TeamHandler) List(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	teams, total, err := h.svc.List(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.List(c, teams, total, limit, offset)
}

func (h *TeamHandler) Create(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	var req model.Team
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

func (h *TeamHandler) Get(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	team, err := h.svc.GetByID(c.Request.Context(), projectID, teamID)
	if err != nil {
		response.NotFound(c, "TEAM")
		return
	}

	response.Success(c, team)
}

func (h *TeamHandler) Update(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	team, err := h.svc.GetByID(c.Request.Context(), projectID, teamID)
	if err != nil {
		response.NotFound(c, "TEAM")
		return
	}

	if err := c.ShouldBindJSON(team); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.Update(c.Request.Context(), team); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, team)
}

func (h *TeamHandler) Delete(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid team id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), projectID, teamID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.NoContent(c)
}

func (h *TeamHandler) GetDefault(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	// Get the first team or create a default one
	team, err := h.svc.GetOrCreateDefault(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, team)
}
