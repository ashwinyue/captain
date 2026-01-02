package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/service"
)

type StaffHandler struct {
	svc           *service.StaffService
	wukongimWSURL string
}

func NewStaffHandler(svc *service.StaffService, wukongimWSURL string) *StaffHandler {
	return &StaffHandler{svc: svc, wukongimWSURL: wukongimWSURL}
}

func (h *StaffHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var projectID *uuid.UUID
	if pid := c.Query("project_id"); pid != "" {
		id, _ := uuid.Parse(pid)
		projectID = &id
	}

	staffs, total, err := h.svc.List(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": staffs, "total": total})
}

func (h *StaffHandler) Get(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	staff, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "staff not found"})
		return
	}
	c.JSON(http.StatusOK, staff)
}

func (h *StaffHandler) Create(c *gin.Context) {
	var staff model.Staff
	if err := c.ShouldBindJSON(&staff); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Create(c.Request.Context(), &staff); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, staff)
}

func (h *StaffHandler) Update(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	staff, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "staff not found"})
		return
	}
	if err := c.ShouldBindJSON(staff); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Update(c.Request.Context(), staff); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, staff)
}

func (h *StaffHandler) Delete(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetMe returns the current staff member
func (h *StaffHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	staff, err := h.svc.GetByID(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "staff not found"})
		return
	}
	c.JSON(http.StatusOK, staff)
}

// UpdateMyServicePaused toggles the current user's service paused status
func (h *StaffHandler) UpdateMyServicePaused(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	var req struct {
		ServicePaused bool `json:"service_paused"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	staff, err := h.svc.UpdateServicePaused(c.Request.Context(), userID.(uuid.UUID), req.ServicePaused)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, staff)
}

// UpdateMyIsActive toggles the current user's active status
func (h *StaffHandler) UpdateMyIsActive(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	// Use query parameter 'active' like Python original
	active := c.Query("active") == "true"
	staff, err := h.svc.UpdateIsActive(c.Request.Context(), userID.(uuid.UUID), active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, staff)
}

// UpdateServicePaused sets a staff member's service paused status
func (h *StaffHandler) UpdateServicePaused(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var req struct {
		ServicePaused bool `json:"service_paused"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	staff, err := h.svc.UpdateServicePaused(c.Request.Context(), id, req.ServicePaused)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, staff)
}

// UpdateIsActive sets a staff member's active status
func (h *StaffHandler) UpdateIsActive(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	// Use query parameter 'active' like Python original
	active := c.Query("active") == "true"
	staff, err := h.svc.UpdateIsActive(c.Request.Context(), id, active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, staff)
}

// GetWuKongIMStatus returns WuKongIM integration status
func (h *StaffHandler) GetWuKongIMStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"connected":     h.wukongimWSURL != "",
		"api_url":       "", // Not needed for frontend
		"websocket_url": h.wukongimWSURL,
	})
}

// CheckWuKongIMOnlineStatus checks staff online status
func (h *StaffHandler) CheckWuKongIMOnlineStatus(c *gin.Context) {
	var req struct {
		StaffIDs []string `json:"staff_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Return all as offline for now
	statuses := make(map[string]bool)
	for _, id := range req.StaffIDs {
		statuses[id] = false
	}
	c.JSON(http.StatusOK, gin.H{"statuses": statuses})
}
