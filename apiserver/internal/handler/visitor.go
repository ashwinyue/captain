package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/service"
)

type VisitorHandler struct {
	svc    *service.VisitorService
	tagSvc *service.TagService
}

func NewVisitorHandler(svc *service.VisitorService, tagSvc *service.TagService) *VisitorHandler {
	return &VisitorHandler{svc: svc, tagSvc: tagSvc}
}

func (h *VisitorHandler) List(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	visitors, total, err := h.svc.List(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": visitors, "total": total})
}

func (h *VisitorHandler) Get(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	visitor, err := h.svc.GetByID(c.Request.Context(), projectID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "visitor not found"})
		return
	}
	c.JSON(http.StatusOK, visitor)
}

func (h *VisitorHandler) Create(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))

	var visitor model.Visitor
	if err := c.ShouldBindJSON(&visitor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	visitor.ProjectID = projectID

	if err := h.svc.Create(c.Request.Context(), &visitor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, visitor)
}

func (h *VisitorHandler) Update(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	visitor, err := h.svc.GetByID(c.Request.Context(), projectID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "visitor not found"})
		return
	}
	if err := c.ShouldBindJSON(visitor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Update(c.Request.Context(), visitor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, visitor)
}

func (h *VisitorHandler) Delete(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	if err := h.svc.Delete(c.Request.Context(), projectID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *VisitorHandler) Block(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	if err := h.svc.Block(c.Request.Context(), projectID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *VisitorHandler) Unblock(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	if err := h.svc.Unblock(c.Request.Context(), projectID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *VisitorHandler) GetTags(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	tags, err := h.tagSvc.GetVisitorTags(c.Request.Context(), projectID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tags})
}

func (h *VisitorHandler) AddTag(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	visitorID, _ := uuid.Parse(c.Param("id"))
	tagID, _ := uuid.Parse(c.Param("tag_id"))

	if err := h.tagSvc.AddToVisitor(c.Request.Context(), projectID, visitorID, tagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *VisitorHandler) RemoveTag(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	visitorID, _ := uuid.Parse(c.Param("id"))
	tagID, _ := uuid.Parse(c.Param("tag_id"))

	if err := h.tagSvc.RemoveFromVisitor(c.Request.Context(), projectID, visitorID, tagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Register handles visitor registration via platform API key (public endpoint)
func (h *VisitorHandler) Register(c *gin.Context) {
	var req service.VisitorRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.PlatformAPIKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid platform API key"})
		return
	}

	// Get client IP from headers
	clientIP := c.ClientIP()
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			clientIP = strings.TrimSpace(parts[0])
		}
	}

	resp, err := h.svc.Register(c.Request.Context(), &req, clientIP)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid platform API key"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetByChannel gets visitor by channel_id
func (h *VisitorHandler) GetByChannel(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	channelID := c.Query("channel_id")

	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id is required"})
		return
	}

	visitor, err := h.svc.GetByChannelID(c.Request.Context(), projectID, channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "visitor not found"})
		return
	}
	c.JSON(http.StatusOK, visitor)
}

// GetBasic gets visitor basic info
func (h *VisitorHandler) GetBasic(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	visitor, err := h.svc.GetByID(c.Request.Context(), projectID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "visitor not found"})
		return
	}

	// Return basic fields only
	c.JSON(http.StatusOK, gin.H{
		"id":                visitor.ID,
		"name":              visitor.Name,
		"nickname":          visitor.Nickname,
		"nickname_zh":       visitor.NicknameZh,
		"avatar_url":        visitor.AvatarURL,
		"email":             visitor.Email,
		"phone_number":      visitor.PhoneNumber,
		"is_online":         visitor.IsOnline,
		"ai_enabled":        visitor.AIEnabled,
		"service_status":    visitor.ServiceStatus,
		"assigned_staff_id": visitor.AssignedStaffID,
		"last_visit_time":   visitor.LastVisitTime,
	})
}

// SetAttributes sets visitor attributes
func (h *VisitorHandler) SetAttributes(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	var attrs map[string]interface{}
	if err := c.ShouldBindJSON(&attrs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	visitor, err := h.svc.SetAttributes(c.Request.Context(), projectID, id, attrs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, visitor)
}

// Accept accepts a visitor (assigns to current staff)
func (h *VisitorHandler) Accept(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	visitorID, _ := uuid.Parse(c.Param("id"))
	staffID, _ := c.Get("user_id")

	resp, err := h.svc.AcceptVisitor(c.Request.Context(), projectID, visitorID, staffID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// EnableAI enables AI for visitor
func (h *VisitorHandler) EnableAI(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	visitor, err := h.svc.SetAIEnabled(c.Request.Context(), projectID, id, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, visitor)
}

// DisableAI disables AI for visitor
func (h *VisitorHandler) DisableAI(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	id, _ := uuid.Parse(c.Param("id"))

	visitor, err := h.svc.SetAIEnabled(c.Request.Context(), projectID, id, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, visitor)
}

// SyncMessages syncs visitor messages (public endpoint)
func (h *VisitorHandler) SyncMessages(c *gin.Context) {
	var req struct {
		PlatformAPIKey  string `json:"platform_api_key"`
		ChannelID       string `json:"channel_id"`
		ChannelType     int    `json:"channel_type"`
		StartMessageSeq *int64 `json:"start_message_seq"`
		EndMessageSeq   *int64 `json:"end_message_seq"`
		Limit           int    `json:"limit"`
		PullMode        int    `json:"pull_mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 100
	}

	resp, err := h.svc.SyncMessages(c.Request.Context(), req.PlatformAPIKey, req.ChannelID, req.ChannelType, req.StartMessageSeq, req.EndMessageSeq, req.Limit, req.PullMode)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid platform API key"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
