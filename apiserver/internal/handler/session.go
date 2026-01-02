package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/service"
)

type SessionHandler struct {
	svc *service.SessionService
}

func NewSessionHandler(svc *service.SessionService) *SessionHandler {
	return &SessionHandler{svc: svc}
}

func (h *SessionHandler) List(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var status *model.SessionStatus
	if s := c.Query("status"); s != "" {
		st := model.SessionStatus(s)
		status = &st
	}

	sessions, total, err := h.svc.List(c.Request.Context(), projectID, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sessions, "total": total})
}

func (h *SessionHandler) Get(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	session, err := h.svc.GetByID(c.Request.Context(), projectID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *SessionHandler) Create(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))

	var session model.Session
	if err := c.ShouldBindJSON(&session); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session.ProjectID = projectID

	if err := h.svc.Create(c.Request.Context(), &session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, session)
}

func (h *SessionHandler) Update(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	session, err := h.svc.GetByID(c.Request.Context(), projectID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if err := c.ShouldBindJSON(session); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Update(c.Request.Context(), session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *SessionHandler) Close(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	if err := h.svc.Close(c.Request.Context(), projectID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *SessionHandler) Transfer(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	var req struct {
		ToStaffID uuid.UUID `json:"to_staff_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Transfer(c.Request.Context(), projectID, id, req.ToStaffID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
