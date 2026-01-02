package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/cloudwego/eino/adk"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/aicenter/internal/config"
	"github.com/tgo/captain/aicenter/internal/pkg/response"
	"github.com/tgo/captain/aicenter/internal/service"
)

type ChatHandler struct {
	runtimeSvc *service.RuntimeService
	cfg        *config.Config
}

func NewChatHandler(runtimeSvc *service.RuntimeService, cfg *config.Config) *ChatHandler {
	return &ChatHandler{runtimeSvc: runtimeSvc, cfg: cfg}
}

// SupervisorConfig matches tgo-ai Python API
type SupervisorConfig struct {
	ExecutionStrategy string `json:"execution_strategy,omitempty"` // single, multiple, auto
	MaxAgents         int    `json:"max_agents,omitempty"`
	Timeout           int    `json:"timeout,omitempty"`
	RequireConsensus  bool   `json:"require_consensus,omitempty"`
}

// SupervisorRunRequest matches tgo-ai Python API
type SupervisorRunRequest struct {
	TeamID         *string           `json:"team_id"`
	AgentID        *string           `json:"agent_id"`
	AgentIDs       []string          `json:"agent_ids"`
	Message        string            `json:"message" binding:"required,min=1,max=10000"`
	SystemMessage  *string           `json:"system_message"`
	ExpectedOutput *string           `json:"expected_output"`
	SessionID      *string           `json:"session_id"`
	UserID         *string           `json:"user_id"`
	Config         *SupervisorConfig `json:"config"`
	Stream         *bool             `json:"stream"` // 默认为 true（流式输出）
	MCPURL         *string           `json:"mcp_url"`
	RAGURL         *string           `json:"rag_url"`
	EnableMemory   bool              `json:"enable_memory"`
}

// Run executes the agent with SSE streaming by default
func (h *ChatHandler) Run(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}

	var req SupervisorRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 默认使用流式输出，除非显式设置 stream=false
	useStream := req.Stream == nil || *req.Stream

	if useStream {
		h.runStream(c, projectID, &req)
	} else {
		h.runSync(c, projectID, &req)
	}
}

func (h *ChatHandler) runSync(c *gin.Context, projectID uuid.UUID, req *SupervisorRunRequest) {
	var resp *service.RunResponse
	var err error

	// Debug routing decision
	log.Printf("[ChatHandler] Routing: agent_id=%v, team_id=%v, agent_ids=%v",
		req.AgentID != nil && *req.AgentID != "",
		req.TeamID != nil && *req.TeamID != "",
		len(req.AgentIDs) > 0)

	// If agent_id is specified, use RunWithAgentTools for direct RAG tool access
	if req.AgentID != nil && *req.AgentID != "" {
		sessionID := ""
		if req.SessionID != nil {
			sessionID = *req.SessionID
		}
		resp, err = h.runtimeSvc.RunWithAgentTools(c.Request.Context(), projectID, *req.AgentID, req.Message, sessionID, req.EnableMemory)
	} else if (req.TeamID == nil || *req.TeamID == "") && len(req.AgentIDs) == 0 {
		// No agent_id, team_id or agent_ids specified - use QueryAnalyzer for smart routing
		log.Printf("[ChatHandler] Using QueryAnalyzer path")
		resp, err = h.runtimeSvc.RunWithQueryAnalyzer(c.Request.Context(), projectID, req.Message)
	} else {
		// Use traditional supervisor routing
		svcReq := &service.RunRequest{
			TeamID:       req.TeamID,
			AgentID:      req.AgentID,
			AgentIDs:     req.AgentIDs,
			Message:      req.Message,
			SessionID:    req.SessionID,
			Stream:       false,
			EnableMemory: req.EnableMemory,
		}
		resp, err = h.runtimeSvc.Run(c.Request.Context(), projectID, svcReq)
	}

	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *ChatHandler) runStream(c *gin.Context, projectID uuid.UUID, req *SupervisorRunRequest) {
	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	svcReq := &service.RunRequest{
		TeamID:       req.TeamID,
		AgentID:      req.AgentID,
		AgentIDs:     req.AgentIDs,
		Message:      req.Message,
		SessionID:    req.SessionID,
		Stream:       true,
		EnableMemory: req.EnableMemory,
	}

	// Send connected event
	h.sendSSE(c, "connected", map[string]string{"status": "connected"})

	err := h.runtimeSvc.Stream(c.Request.Context(), projectID, svcReq, func(event *adk.AgentEvent) error {
		sseEvent := service.ConvertADKEvent(event)
		return h.sendSSE(c, "event", sseEvent)
	})

	if err != nil {
		h.sendSSE(c, "error", map[string]string{"error": err.Error()})
	}

	// Send done event
	h.sendSSE(c, "done", map[string]string{"status": "done"})
}

func (h *ChatHandler) sendSSE(c *gin.Context, eventType string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	c.SSEvent(eventType, string(jsonData))
	c.Writer.Flush()
	return nil
}

// Cancel cancels a running agent execution
func (h *ChatHandler) Cancel(c *gin.Context) {
	runID := c.Param("run_id")

	cancelled, reason := h.runtimeSvc.Cancel(c.Request.Context(), runID)

	c.JSON(http.StatusAccepted, gin.H{
		"run_id":    runID,
		"cancelled": cancelled,
		"reason":    reason,
	})
}

// CompletionsRequest represents OpenAI-compatible chat completions request
type CompletionsRequest struct {
	Model       string              `json:"model" binding:"required"`
	Messages    []CompletionMessage `json:"messages" binding:"required"`
	Stream      bool                `json:"stream"`
	Temperature *float64            `json:"temperature,omitempty"`
	MaxTokens   *int                `json:"max_tokens,omitempty"`
	TopP        *float64            `json:"top_p,omitempty"`
	AgentID     *string             `json:"agent_id,omitempty"` // Optional agent ID for RAG tools
}

type CompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Completions provides OpenAI-compatible chat completions API
func (h *ChatHandler) Completions(c *gin.Context) {
	projectID, err := uuid.Parse(c.GetString("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "invalid project_id",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	var req CompletionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	// Extract last user message
	var userMessage string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			userMessage = req.Messages[i].Content
			break
		}
	}

	if userMessage == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "no user message found",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	svcReq := &service.RunRequest{
		AgentID: req.AgentID,
		Message: userMessage,
		Stream:  req.Stream,
	}

	if req.Stream {
		// SSE streaming for OpenAI-compatible format
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

		runID := uuid.New().String()

		err := h.runtimeSvc.Stream(c.Request.Context(), projectID, svcReq, func(event *adk.AgentEvent) error {
			content := service.GetEventContent(event)
			if content == "" {
				return nil
			}
			chunk := gin.H{
				"id":     runID,
				"object": "chat.completion.chunk",
				"model":  req.Model,
				"choices": []gin.H{{
					"index": 0,
					"delta": gin.H{
						"content": content,
					},
					"finish_reason": nil,
				}},
			}
			data, _ := json.Marshal(chunk)
			c.Writer.WriteString("data: " + string(data) + "\n\n")
			c.Writer.Flush()
			return nil
		})

		if err != nil {
			c.Writer.WriteString("data: [DONE]\n\n")
		} else {
			// Send final chunk with finish_reason
			finalChunk := gin.H{
				"id":     runID,
				"object": "chat.completion.chunk",
				"model":  req.Model,
				"choices": []gin.H{{
					"index":         0,
					"delta":         gin.H{},
					"finish_reason": "stop",
				}},
			}
			data, _ := json.Marshal(finalChunk)
			c.Writer.WriteString("data: " + string(data) + "\n\n")
			c.Writer.WriteString("data: [DONE]\n\n")
		}
		c.Writer.Flush()
	} else {
		// Non-streaming response
		var resp *service.RunResponse
		var err error

		// If agent_id is specified, try to use ReAct agent with tools
		if req.AgentID != nil && *req.AgentID != "" {
			resp, err = h.runtimeSvc.RunWithAgentTools(c.Request.Context(), projectID, *req.AgentID, userMessage, "", false)
		} else {
			resp, err = h.runtimeSvc.Run(c.Request.Context(), projectID, svcReq)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"message": err.Error(),
					"type":    "api_error",
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":     uuid.New().String(),
			"object": "chat.completion",
			"model":  req.Model,
			"choices": []gin.H{{
				"index": 0,
				"message": gin.H{
					"role":    "assistant",
					"content": resp.Content,
				},
				"finish_reason": "stop",
			}},
			"usage": gin.H{
				"prompt_tokens":     0,
				"completion_tokens": 0,
				"total_tokens":      0,
			},
		})
	}
}
