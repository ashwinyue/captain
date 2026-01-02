package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/service"
)

type QueueHandler struct {
	svc *service.QueueService
}

func NewQueueHandler(svc *service.QueueService) *QueueHandler {
	return &QueueHandler{svc: svc}
}

func (h *QueueHandler) List(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, total, err := h.svc.ListWaiting(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
}

func (h *QueueHandler) Add(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))

	var item model.VisitorWaitingQueue
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item.ProjectID = projectID

	if err := h.svc.AddToQueue(c.Request.Context(), &item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *QueueHandler) Assign(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	var req struct {
		StaffID uuid.UUID `json:"staff_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.AssignToStaff(c.Request.Context(), projectID, id, req.StaffID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *QueueHandler) Remove(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	if err := h.svc.RemoveFromQueue(c.Request.Context(), projectID, id, model.QueueStatusLeft); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *QueueHandler) GetPosition(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	position, err := h.svc.GetQueuePosition(c.Request.Context(), projectID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"position": position})
}

func (h *QueueHandler) GetCount(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))

	waiting, assigned, err := h.svc.GetQueueCount(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"waiting":  waiting,
		"assigned": assigned,
		"total":    waiting + assigned,
	})
}
