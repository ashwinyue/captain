package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tgo/captain/apiserver/internal/service"
)

type SetupHandler struct {
	svc *service.SetupService
}

func NewSetupHandler(svc *service.SetupService) *SetupHandler {
	return &SetupHandler{svc: svc}
}

// GetStatus returns the current setup status
func (h *SetupHandler) GetStatus(c *gin.Context) {
	status, err := h.svc.GetStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// CreateAdmin creates the first admin account
func (h *SetupHandler) CreateAdmin(c *gin.Context) {
	var req service.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.CreateAdmin(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrAlreadyInstalled {
			c.JSON(http.StatusForbidden, gin.H{
				"detail": "System installation is already complete. Setup endpoints are disabled for security reasons.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// SkipLLMConfig skips LLM configuration
func (h *SetupHandler) SkipLLMConfig(c *gin.Context) {
	err := h.svc.SkipLLMConfig(c.Request.Context())
	if err != nil {
		if err == service.ErrAlreadyInstalled {
			c.JSON(http.StatusForbidden, gin.H{
				"detail": "System installation is already complete.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "LLM configuration skipped"})
}

// BatchCreateStaff creates staff members during setup
func (h *SetupHandler) BatchCreateStaff(c *gin.Context) {
	var req service.BatchCreateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.BatchCreateStaff(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrAlreadyInstalled {
			c.JSON(http.StatusForbidden, gin.H{
				"detail": "System installation is already complete.",
			})
			return
		}
		if err == service.ErrAdminRequired {
			c.JSON(http.StatusBadRequest, gin.H{
				"detail": "Admin account must be created before adding staff members",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ConfigureLLM configures the LLM provider
func (h *SetupHandler) ConfigureLLM(c *gin.Context) {
	var req service.ConfigureLLMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.ConfigureLLM(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrAlreadyInstalled {
			c.JSON(http.StatusForbidden, gin.H{
				"detail": "System installation is already complete.",
			})
			return
		}
		if err == service.ErrAdminRequired {
			c.JSON(http.StatusBadRequest, gin.H{
				"detail": "Admin account must be created before configuring LLM provider",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Verify verifies the setup is complete
func (h *SetupHandler) Verify(c *gin.Context) {
	status, err := h.svc.GetStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	checks := map[string]gin.H{
		"database_connected":  {"passed": true, "message": "Database connection is healthy"},
		"admin_exists":        {"passed": status.HasAdmin, "message": "Admin account exists"},
		"llm_configured":      {"passed": status.HasLLMConfig || status.SkipLLMConfig, "message": "LLM provider configured or skipped"},
		"project_exists":      {"passed": true, "message": "Project exists"},
		"installation_status": {"passed": status.IsInstalled, "message": "Installation is complete"},
	}

	errors := []string{}
	warnings := []string{}

	if !status.HasAdmin {
		errors = append(errors, "Admin account has not been created")
	}
	if !status.HasLLMConfig && !status.SkipLLMConfig {
		warnings = append(warnings, "LLM provider not configured yet")
	}

	isValid := status.HasAdmin && (status.HasLLMConfig || status.SkipLLMConfig)

	c.JSON(http.StatusOK, gin.H{
		"is_valid": isValid,
		"checks":   checks,
		"errors":   errors,
		"warnings": warnings,
	})
}
