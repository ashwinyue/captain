package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/service"
)

type OnboardingHandler struct {
	svc *service.OnboardingService
}

func NewOnboardingHandler(svc *service.OnboardingService) *OnboardingHandler {
	return &OnboardingHandler{svc: svc}
}

// GetProgress returns the onboarding progress
func (h *OnboardingHandler) GetProgress(c *gin.Context) {
	projectIDStr := c.GetString("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	progress, err := h.svc.GetProgress(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// SkipRequest represents the skip request
type SkipRequest struct {
	StepNumber *int `json:"step_number,omitempty"`
}

// Skip skips onboarding step(s)
func (h *OnboardingHandler) Skip(c *gin.Context) {
	projectIDStr := c.GetString("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	var req SkipRequest
	c.ShouldBindJSON(&req)

	progress, err := h.svc.SkipStep(c.Request.Context(), projectID, req.StepNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// Reset resets the onboarding progress
func (h *OnboardingHandler) Reset(c *gin.Context) {
	projectIDStr := c.GetString("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	progress, err := h.svc.Reset(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}
