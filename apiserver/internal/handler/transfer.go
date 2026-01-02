package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/service"
)

type TransferHandler struct {
	svc *service.TransferService
}

func NewTransferHandler(svc *service.TransferService) *TransferHandler {
	return &TransferHandler{svc: svc}
}

// TransferToStaffRequest represents transfer to staff request
type TransferToStaffRequest struct {
	VisitorID     string  `json:"visitor_id" binding:"required"`
	TargetStaffID *string `json:"target_staff_id,omitempty"`
	AIDisabled    *bool   `json:"ai_disabled,omitempty"`
}

// TransferToStaff handles visitor transfer to staff
func (h *TransferHandler) TransferToStaff(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))

	var req TransferToStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	visitorID, err := uuid.Parse(req.VisitorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visitor_id"})
		return
	}

	var targetStaffID *uuid.UUID
	if req.TargetStaffID != nil {
		id, err := uuid.Parse(*req.TargetStaffID)
		if err == nil {
			targetStaffID = &id
		}
	}

	result, err := h.svc.TransferToStaff(c.Request.Context(), &service.TransferRequest{
		VisitorID:           visitorID,
		ProjectID:           projectID,
		Source:              service.AssignmentSourceManual,
		TargetStaffID:       targetStaffID,
		AddToQueueIfNoStaff: true,
		AIDisabled:          req.AIDisabled,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// TransferToStaffByPlatformKey handles visitor transfer to staff via platform API key (public endpoint)
func (h *TransferHandler) TransferToStaffByPlatformKey(c *gin.Context) {
	var req struct {
		APIKey        string  `json:"api_key" binding:"required"`
		VisitorID     string  `json:"visitor_id" binding:"required"`
		TargetStaffID *string `json:"target_staff_id,omitempty"`
		AIDisabled    *bool   `json:"ai_disabled,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	visitorID, err := uuid.Parse(req.VisitorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visitor_id"})
		return
	}

	// Get project ID from platform API key
	projectID, err := h.svc.GetProjectIDByAPIKey(c.Request.Context(), req.APIKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid platform API key"})
		return
	}

	var targetStaffID *uuid.UUID
	if req.TargetStaffID != nil {
		id, err := uuid.Parse(*req.TargetStaffID)
		if err == nil {
			targetStaffID = &id
		}
	}

	result, err := h.svc.TransferToStaff(c.Request.Context(), &service.TransferRequest{
		VisitorID:           visitorID,
		ProjectID:           projectID,
		Source:              service.AssignmentSourceRule,
		TargetStaffID:       targetStaffID,
		AddToQueueIfNoStaff: true,
		AIDisabled:          req.AIDisabled,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
