package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/aicenter/internal/model"
	"github.com/tgo/captain/aicenter/internal/pkg/response"
	"github.com/tgo/captain/aicenter/internal/service"
)

type ProviderHandler struct {
	svc *service.ProviderService
}

func NewProviderHandler(svc *service.ProviderService) *ProviderHandler {
	return &ProviderHandler{svc: svc}
}

func (h *ProviderHandler) List(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	providers, total, err := h.svc.List(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.List(c, providers, total, limit, offset)
}

func (h *ProviderHandler) Create(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	var req model.LLMProvider
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

func (h *ProviderHandler) Get(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	providerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid provider id")
		return
	}

	provider, err := h.svc.GetByID(c.Request.Context(), projectID, providerID)
	if err != nil {
		response.NotFound(c, "LLM_PROVIDER")
		return
	}

	response.Success(c, provider)
}

func (h *ProviderHandler) Update(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	providerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid provider id")
		return
	}

	provider, err := h.svc.GetByID(c.Request.Context(), projectID, providerID)
	if err != nil {
		response.NotFound(c, "LLM_PROVIDER")
		return
	}

	if err := c.ShouldBindJSON(provider); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.Update(c.Request.Context(), provider); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, provider)
}

func (h *ProviderHandler) Delete(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	providerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid provider id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), projectID, providerID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.NoContent(c)
}

type ProviderSyncRequest struct {
	Providers []model.LLMProvider `json:"providers" binding:"required"`
}

func (h *ProviderHandler) Sync(c *gin.Context) {
	var req ProviderSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if len(req.Providers) == 0 {
		response.BadRequest(c, "No providers provided for sync")
		return
	}

	providers, err := h.svc.Sync(c.Request.Context(), req.Providers)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	if len(providers) != len(req.Providers) {
		response.InternalError(c, "Sync incomplete: count mismatch")
		return
	}

	response.Success(c, gin.H{"data": providers})
}

// Test tests the connection to an AI provider
func (h *ProviderHandler) Test(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	providerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid provider id")
		return
	}

	provider, err := h.svc.GetByID(c.Request.Context(), projectID, providerID)
	if err != nil {
		response.NotFound(c, "LLM_PROVIDER")
		return
	}

	// Build test request based on provider kind
	method, url, headers, testErr := buildTestRequest(provider)
	if testErr != nil {
		response.Success(c, gin.H{
			"success": false,
			"message": testErr.Error(),
		})
		return
	}

	// Make test request
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest(method, url, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		response.Success(c, gin.H{
			"success": false,
			"message": fmt.Sprintf("Connection failed: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	ok := resp.StatusCode >= 200 && resp.StatusCode < 300

	response.Success(c, gin.H{
		"success": ok,
		"message": map[bool]string{true: "Connection test passed", false: fmt.Sprintf("HTTP %d", resp.StatusCode)}[ok],
		"status":  resp.StatusCode,
		"details": string(body),
	})
}

func buildTestRequest(provider *model.LLMProvider) (string, string, map[string]string, error) {
	if provider.APIKey == "" {
		return "", "", nil, fmt.Errorf("API key is not set for this provider")
	}

	kind := strings.ToLower(provider.ProviderKind)
	base := strings.TrimRight(provider.APIBaseURL, "/")

	switch kind {
	case "openai", "gpt", "oai":
		if base == "" {
			base = "https://api.openai.com/v1"
		}
		return "GET", base + "/models", map[string]string{"Authorization": "Bearer " + provider.APIKey}, nil

	case "anthropic", "claude":
		if base == "" {
			base = "https://api.anthropic.com"
		}
		version := "2023-06-01"
		if cfg, ok := provider.Config["anthropic_version"].(string); ok && cfg != "" {
			version = cfg
		}
		return "GET", base + "/v1/models", map[string]string{
			"x-api-key":         provider.APIKey,
			"anthropic-version": version,
		}, nil

	case "dashscope", "ali", "aliyun":
		if base == "" || !strings.Contains(base, "compatible-mode") {
			base = "https://dashscope.aliyuncs.com/compatible-mode/v1"
		}
		return "GET", base + "/models", map[string]string{"Authorization": "Bearer " + provider.APIKey}, nil

	case "azure_openai", "azure-openai", "azure":
		if base == "" {
			return "", "", nil, fmt.Errorf("api_base_url is required for Azure OpenAI")
		}
		root := base
		if !strings.Contains(base, "/openai") {
			root = base + "/openai"
		}
		apiVersion := "2023-12-01-preview"
		if cfg, ok := provider.Config["api_version"].(string); ok && cfg != "" {
			apiVersion = cfg
		}
		return "GET", fmt.Sprintf("%s/deployments?api-version=%s", root, apiVersion), map[string]string{"api-key": provider.APIKey}, nil

	default:
		// OpenAI-compatible fallback
		if base != "" {
			return "GET", base + "/models", map[string]string{"Authorization": "Bearer " + provider.APIKey}, nil
		}
		return "", "", nil, fmt.Errorf("unsupported provider: %s", provider.ProviderKind)
	}
}
