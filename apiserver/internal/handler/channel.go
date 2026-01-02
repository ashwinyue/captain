package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/service"
)

type ChannelHandler struct {
	svc *service.ChannelService
}

func NewChannelHandler(svc *service.ChannelService) *ChannelHandler {
	return &ChannelHandler{svc: svc}
}

func (h *ChannelHandler) List(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	channels, total, err := h.svc.List(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": channels, "total": total})
}

func (h *ChannelHandler) Get(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	channelID := c.Param("id")

	channel, err := h.svc.GetByID(c.Request.Context(), projectID, channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}
	c.JSON(http.StatusOK, channel)
}

func (h *ChannelHandler) Create(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))

	var req struct {
		model.Channel
		Subscribers []string `json:"subscribers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Channel.ProjectID = projectID

	if err := h.svc.Create(c.Request.Context(), &req.Channel, req.Subscribers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req.Channel)
}

func (h *ChannelHandler) Delete(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	channelID := c.Param("id")

	if err := h.svc.Delete(c.Request.Context(), projectID, channelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *ChannelHandler) AddMembers(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	channelID := c.Param("id")

	var req struct {
		UIDs []string `json:"uids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.AddMembers(c.Request.Context(), projectID, channelID, req.UIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *ChannelHandler) RemoveMembers(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	channelID := c.Param("id")

	var req struct {
		UIDs []string `json:"uids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.RemoveMembers(c.Request.Context(), projectID, channelID, req.UIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
