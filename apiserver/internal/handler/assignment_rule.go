package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/service"
)

type AssignmentRuleHandler struct {
	svc *service.AssignmentRuleService
}

func NewAssignmentRuleHandler(svc *service.AssignmentRuleService) *AssignmentRuleHandler {
	return &AssignmentRuleHandler{svc: svc}
}

func (h *AssignmentRuleHandler) Get(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))

	rule, err := h.svc.Get(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add effective_prompt
	effectivePrompt := rule.Prompt
	if effectivePrompt == "" {
		effectivePrompt = h.svc.GetDefaultPrompt()
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                     rule.ID,
		"project_id":             rule.ProjectID,
		"ai_provider_id":         rule.AIProviderID,
		"model":                  rule.Model,
		"prompt":                 rule.Prompt,
		"effective_prompt":       effectivePrompt,
		"llm_assignment_enabled": rule.LLMAssignmentEnabled,
		"timezone":               rule.Timezone,
		"service_weekdays":       rule.ServiceWeekdays,
		"service_start_time":     rule.ServiceStartTime,
		"service_end_time":       rule.ServiceEndTime,
		"max_concurrent_chats":   rule.MaxConcurrentChats,
		"auto_close_hours":       rule.AutoCloseHours,
		"created_at":             rule.CreatedAt,
		"updated_at":             rule.UpdatedAt,
	})
}

func (h *AssignmentRuleHandler) Update(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule, err := h.svc.Upsert(c.Request.Context(), projectID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	effectivePrompt := rule.Prompt
	if effectivePrompt == "" {
		effectivePrompt = h.svc.GetDefaultPrompt()
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                     rule.ID,
		"project_id":             rule.ProjectID,
		"ai_provider_id":         rule.AIProviderID,
		"model":                  rule.Model,
		"prompt":                 rule.Prompt,
		"effective_prompt":       effectivePrompt,
		"llm_assignment_enabled": rule.LLMAssignmentEnabled,
		"timezone":               rule.Timezone,
		"service_weekdays":       rule.ServiceWeekdays,
		"service_start_time":     rule.ServiceStartTime,
		"service_end_time":       rule.ServiceEndTime,
		"max_concurrent_chats":   rule.MaxConcurrentChats,
		"auto_close_hours":       rule.AutoCloseHours,
		"created_at":             rule.CreatedAt,
		"updated_at":             rule.UpdatedAt,
	})
}

func (h *AssignmentRuleHandler) GetDefaultPrompt(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"default_prompt": h.svc.GetDefaultPrompt(),
	})
}
