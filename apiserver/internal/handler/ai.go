package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/pkg/aicenter"
	"github.com/tgo/captain/apiserver/internal/pkg/wukongim"
)

type AIHandler struct {
	client   *aicenter.Client
	imClient *wukongim.Client
}

func NewAIHandler(client *aicenter.Client, imClient *wukongim.Client) *AIHandler {
	return &AIHandler{client: client, imClient: imClient}
}

func (h *AIHandler) getHeaders(c *gin.Context) map[string]string {
	headers := map[string]string{}
	if auth := c.GetHeader("Authorization"); auth != "" {
		headers["Authorization"] = auth
	}
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		headers["X-API-Key"] = apiKey
	}
	// Get project ID from header or context
	if projectID := c.GetHeader("X-Project-ID"); projectID != "" {
		headers["X-Project-ID"] = projectID
	} else if projectID := c.GetString("project_id"); projectID != "" {
		headers["X-Project-ID"] = projectID
	}
	return headers
}

func (h *AIHandler) respond(c *gin.Context, data []byte, statusCode int, err error) {
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.Data(statusCode, "application/json", data)
}

// Agents

func (h *AIHandler) ListAgents(c *gin.Context) {
	projectID := c.GetString("project_id")
	data, status, err := h.client.ListAgents(c.Request.Context(), projectID, h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) GetAgent(c *gin.Context) {
	data, status, err := h.client.GetAgent(c.Request.Context(), c.Param("id"), h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) CreateAgent(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body["project_id"] = c.GetString("project_id")
	data, status, err := h.client.CreateAgent(c.Request.Context(), body, h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) UpdateAgent(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.UpdateAgent(c.Request.Context(), c.Param("id"), body, h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) DeleteAgent(c *gin.Context) {
	data, status, err := h.client.DeleteAgent(c.Request.Context(), c.Param("id"), h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) RunAgent(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.RunAgent(c.Request.Context(), c.Param("id"), body, h.getHeaders(c))
	h.respond(c, data, status, err)
}

// Teams

func (h *AIHandler) ListTeams(c *gin.Context) {
	projectID := c.GetString("project_id")
	data, status, err := h.client.ListTeams(c.Request.Context(), projectID, h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) GetTeam(c *gin.Context) {
	data, status, err := h.client.GetTeam(c.Request.Context(), c.Param("id"), h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) CreateTeam(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body["project_id"] = c.GetString("project_id")
	data, status, err := h.client.CreateTeam(c.Request.Context(), body, h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) UpdateTeam(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.UpdateTeam(c.Request.Context(), c.Param("id"), body, h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) DeleteTeam(c *gin.Context) {
	data, status, err := h.client.DeleteTeam(c.Request.Context(), c.Param("id"), h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) GetDefaultTeam(c *gin.Context) {
	// Get default team - returns the first team or creates one if none exists
	data, status, err := h.client.GetDefaultTeam(c.Request.Context(), c.GetString("project_id"), h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) RunTeam(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.RunTeam(c.Request.Context(), c.Param("id"), body, h.getHeaders(c))
	h.respond(c, data, status, err)
}

// Tools

func (h *AIHandler) ListTools(c *gin.Context) {
	projectID := c.GetString("project_id")
	data, status, err := h.client.ListTools(c.Request.Context(), projectID, h.getHeaders(c))
	// Extract just the data array from paginated response for Python API compatibility
	h.respondDataArray(c, data, status, err)
}

// respondDataArray extracts data array from paginated response {data: [], pagination: {}}
func (h *AIHandler) respondDataArray(c *gin.Context, data []byte, status int, err error) {
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if status >= 400 {
		c.Data(status, "application/json", data)
		return
	}
	// Try to extract "data" field from response
	var resp map[string]interface{}
	if json.Unmarshal(data, &resp) == nil {
		if dataArr, ok := resp["data"]; ok {
			c.JSON(status, dataArr)
			return
		}
	}
	// Fallback to original response
	c.Data(status, "application/json", data)
}

func (h *AIHandler) GetTool(c *gin.Context) {
	data, status, err := h.client.GetTool(c.Request.Context(), c.Param("id"), h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) CreateTool(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body["project_id"] = c.GetString("project_id")
	data, status, err := h.client.CreateTool(c.Request.Context(), body, h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) UpdateTool(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.UpdateTool(c.Request.Context(), c.Param("id"), body, h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) DeleteTool(c *gin.Context) {
	data, status, err := h.client.DeleteTool(c.Request.Context(), c.Param("id"), h.getHeaders(c))
	h.respond(c, data, status, err)
}

// Providers

func (h *AIHandler) ListProviders(c *gin.Context) {
	projectID := c.GetString("project_id")
	data, status, err := h.client.ListProviders(c.Request.Context(), projectID, h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) GetProvider(c *gin.Context) {
	data, status, err := h.client.GetProvider(c.Request.Context(), c.Param("id"), h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) CreateProvider(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body["project_id"] = c.GetString("project_id")
	data, status, err := h.client.CreateProvider(c.Request.Context(), body, h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) UpdateProvider(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.UpdateProvider(c.Request.Context(), c.Param("id"), body, h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) DeleteProvider(c *gin.Context) {
	data, status, err := h.client.DeleteProvider(c.Request.Context(), c.Param("id"), h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) EnableProvider(c *gin.Context) {
	// Stub: aicenter doesn't have this endpoint yet
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Provider enabled"})
}

func (h *AIHandler) DisableProvider(c *gin.Context) {
	// Stub: aicenter doesn't have this endpoint yet
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Provider disabled"})
}

func (h *AIHandler) SyncProvider(c *gin.Context) {
	data, status, err := h.client.SyncProvider(c.Request.Context(), c.Param("id"), h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) TestProvider(c *gin.Context) {
	data, status, err := h.client.TestProvider(c.Request.Context(), c.Param("id"), h.getHeaders(c))
	h.respond(c, data, status, err)
}

// Models

func (h *AIHandler) ListModels(c *gin.Context) {
	projectID := c.GetString("project_id")
	data, status, err := h.client.ListModels(c.Request.Context(), projectID, h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) GetModel(c *gin.Context) {
	data, status, err := h.client.GetModel(c.Request.Context(), c.Param("id"), h.getHeaders(c))
	h.respond(c, data, status, err)
}

// FetchModels fetches available models from a provider API
func (h *AIHandler) FetchModels(c *gin.Context) {
	var req struct {
		Provider   string                 `json:"provider"`
		APIKey     string                 `json:"api_key,omitempty"`
		APIBaseURL string                 `json:"api_base_url,omitempty"`
		Config     map[string]interface{} `json:"config,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return default models based on provider
	models := getDefaultModels(req.Provider)
	c.JSON(http.StatusOK, gin.H{
		"provider":    req.Provider,
		"models":      models,
		"is_fallback": true,
	})
}

// getDefaultModels returns default models for a provider
func getDefaultModels(provider string) []map[string]interface{} {
	defaults := map[string][]map[string]interface{}{
		"openai": {
			{"id": "gpt-4o", "name": "GPT-4o", "model_type": "chat"},
			{"id": "gpt-4o-mini", "name": "GPT-4o Mini", "model_type": "chat"},
			{"id": "gpt-4-turbo", "name": "GPT-4 Turbo", "model_type": "chat"},
			{"id": "gpt-3.5-turbo", "name": "GPT-3.5 Turbo", "model_type": "chat"},
			{"id": "text-embedding-3-large", "name": "Text Embedding 3 Large", "model_type": "embedding"},
			{"id": "text-embedding-3-small", "name": "Text Embedding 3 Small", "model_type": "embedding"},
		},
		"anthropic": {
			{"id": "claude-sonnet-4-20250514", "name": "Claude Sonnet 4", "model_type": "chat"},
			{"id": "claude-3-5-sonnet-20241022", "name": "Claude 3.5 Sonnet", "model_type": "chat"},
			{"id": "claude-3-5-haiku-20241022", "name": "Claude 3.5 Haiku", "model_type": "chat"},
			{"id": "claude-3-opus-20240229", "name": "Claude 3 Opus", "model_type": "chat"},
		},
		"dashscope": {
			{"id": "qwen-max", "name": "Qwen Max", "model_type": "chat"},
			{"id": "qwen-plus", "name": "Qwen Plus", "model_type": "chat"},
			{"id": "qwen-turbo", "name": "Qwen Turbo", "model_type": "chat"},
			{"id": "text-embedding-v3", "name": "Text Embedding V3", "model_type": "embedding"},
		},
	}

	if models, ok := defaults[provider]; ok {
		return models
	}
	// Default fallback
	return []map[string]interface{}{
		{"id": "gpt-4o", "name": "GPT-4o", "model_type": "chat"},
		{"id": "gpt-4o-mini", "name": "GPT-4o Mini", "model_type": "chat"},
	}
}

// Project AI Configs

func (h *AIHandler) GetProjectAIConfig(c *gin.Context) {
	projectID := c.GetString("project_id")
	data, status, err := h.client.GetProjectAIConfig(c.Request.Context(), projectID, h.getHeaders(c))
	h.respond(c, data, status, err)
}

func (h *AIHandler) UpsertProjectAIConfig(c *gin.Context) {
	projectID := c.GetString("project_id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, status, err := h.client.UpsertProjectAIConfig(c.Request.Context(), projectID, body, h.getHeaders(c))
	h.respond(c, data, status, err)
}

// GetProjectAIConfigByID gets project AI config using project ID from URL parameter
func (h *AIHandler) GetProjectAIConfigByID(c *gin.Context) {
	projectID := c.Param("id")
	data, status, err := h.client.GetProjectAIConfig(c.Request.Context(), projectID, h.getHeaders(c))
	// Return empty config if not found (matches Python behavior)
	if status == http.StatusNotFound {
		c.JSON(http.StatusOK, gin.H{
			"project_id":                    projectID,
			"default_chat_provider_id":      nil,
			"default_chat_model":            nil,
			"default_embedding_provider_id": nil,
			"default_embedding_model":       nil,
		})
		return
	}
	h.respond(c, data, status, err)
}

// UpsertProjectAIConfigByID upserts project AI config using project ID from URL parameter
func (h *AIHandler) UpsertProjectAIConfigByID(c *gin.Context) {
	projectID := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Add project_id to body for aicenter
	body["project_id"] = projectID
	data, status, err := h.client.UpsertProjectAIConfig(c.Request.Context(), projectID, body, h.getHeaders(c))
	h.respond(c, data, status, err)
}

// TeamChatRequest represents the request body for team chat
type TeamChatRequest struct {
	AgentID        string  `json:"agent_id"`
	TeamID         string  `json:"team_id"`
	Message        string  `json:"message"`
	SystemMessage  *string `json:"system_message"`
	ExpectedOutput *string `json:"expected_output"`
	EnableMemory   bool    `json:"enable_memory"`
}

// TeamChatResponse represents the response from aicenter
type TeamChatResponse struct {
	Content string `json:"content"`
	RunID   string `json:"run_id"`
}

// TeamChat handles staff chat with AI team or agent
func (h *AIHandler) TeamChat(c *gin.Context) {
	projectID := c.GetString("project_id")
	// user_id is stored as uuid.UUID, not string
	userIDValue, _ := c.Get("user_id")
	userID := ""
	if uid, ok := userIDValue.(uuid.UUID); ok {
		userID = uid.String()
	}

	var req TeamChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determine channel_id
	var channelID string
	if req.TeamID != "" {
		channelID = req.TeamID + "-team"
	} else {
		channelID = req.AgentID + "-agent"
	}
	staffUID := userID + "-staff"

	// 1. Send staff message to WuKongIM first (to create/update conversation)
	staffMsgNo := uuid.New().String()
	if h.imClient != nil && req.Message != "" {
		_, _ = h.imClient.SendTextMessage(c.Request.Context(), &wukongim.SendTextMessageRequest{
			FromUID:     staffUID,
			ChannelID:   channelID,
			ChannelType: 1,
			Content:     req.Message,
			ClientMsgNo: staffMsgNo,
		})
	}

	// 2. Build body for aicenter (enable_memory always true like Python original)
	// session_id format: {channel_id}@{channel_type} - used for memory tracking
	sessionID := fmt.Sprintf("%s@%d", channelID, 1)
	body := map[string]interface{}{
		"project_id":    projectID,
		"message":       req.Message,
		"enable_memory": true,
		"session_id":    sessionID,
		"user_id":       userID,
		"stream":        false, // TeamChat 需要非流式响应
	}
	if req.AgentID != "" {
		body["agent_id"] = req.AgentID
	}
	if req.TeamID != "" {
		body["team_id"] = req.TeamID
	}
	if req.SystemMessage != nil {
		body["system_message"] = *req.SystemMessage
	}
	if req.ExpectedOutput != nil {
		body["expected_output"] = *req.ExpectedOutput
	}

	// 3. Call aicenter
	data, status, err := h.client.TeamChat(c.Request.Context(), body, h.getHeaders(c))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if status != http.StatusOK {
		c.Data(status, "application/json", data)
		return
	}

	// 4. Parse response
	var resp TeamChatResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse AI response"})
		return
	}

	// 5. Send AI response to WuKongIM
	aiMsgNo := uuid.New().String()
	if h.imClient != nil && resp.Content != "" {
		_, _ = h.imClient.SendTextMessage(c.Request.Context(), &wukongim.SendTextMessageRequest{
			FromUID:     channelID,
			ChannelID:   staffUID,
			ChannelType: 1,
			Content:     resp.Content,
			ClientMsgNo: aiMsgNo,
		})
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Request processed",
		"client_msg_no": aiMsgNo,
		"content":       resp.Content,
	})
}

// TeamChatStreamRequest 流式对话请求
type TeamChatStreamRequest struct {
	AgentID        string  `json:"agent_id"`
	TeamID         string  `json:"team_id"`
	Message        string  `json:"message" binding:"required"`
	SystemMessage  *string `json:"system_message"`
	ExpectedOutput *string `json:"expected_output"`
	SessionID      *string `json:"session_id"`
	EnableMemory   bool    `json:"enable_memory"`
}

// TeamChatStream handles streaming chat with AI team or agent (SSE proxy)
func (h *AIHandler) TeamChatStream(c *gin.Context) {
	projectID := c.GetString("project_id")
	userIDValue, _ := c.Get("user_id")
	userID := ""
	if uid, ok := userIDValue.(uuid.UUID); ok {
		userID = uid.String()
	}

	var req TeamChatStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build body for aicenter
	var channelID string
	if req.TeamID != "" {
		channelID = req.TeamID + "-team"
	} else {
		channelID = req.AgentID + "-agent"
	}
	sessionID := fmt.Sprintf("%s@%d", channelID, 1)

	body := map[string]interface{}{
		"project_id":    projectID,
		"message":       req.Message,
		"enable_memory": req.EnableMemory,
		"session_id":    sessionID,
		"user_id":       userID,
		"stream":        true, // 强制流式
	}
	if req.AgentID != "" {
		body["agent_id"] = req.AgentID
	}
	if req.TeamID != "" {
		body["team_id"] = req.TeamID
	}
	if req.SystemMessage != nil {
		body["system_message"] = *req.SystemMessage
	}
	if req.ExpectedOutput != nil {
		body["expected_output"] = *req.ExpectedOutput
	}

	// Call aicenter with streaming
	resp, err := h.client.TeamChatStream(c.Request.Context(), body, h.getHeaders(c))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	// Proxy the SSE stream
	c.Stream(func(w io.Writer) bool {
		buf := make([]byte, 1024)
		n, err := resp.Body.Read(buf)
		if err != nil {
			return false
		}
		if n > 0 {
			w.Write(buf[:n])
			c.Writer.Flush()
		}
		return true
	})
}
