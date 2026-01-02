package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/pkg/wukongim"
	"github.com/tgo/captain/apiserver/internal/service"
)

type ChatHandler struct {
	svc *service.ChatService
}

func NewChatHandler(svc *service.ChatService) *ChatHandler {
	return &ChatHandler{svc: svc}
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))

	var req service.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, err := h.svc.SendMessage(c.Request.Context(), projectID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, msg)
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	channelID := c.Query("channel_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, total, err := h.svc.GetMessages(c.Request.Context(), projectID, channelID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": messages, "total": total})
}

func (h *ChatHandler) RevokeMessage(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	messageID := c.Param("id")

	if err := h.svc.RevokeMessage(c.Request.Context(), projectID, messageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *ChatHandler) GetConversations(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	uid := c.Query("uid")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	conversations, total, err := h.svc.GetConversations(c.Request.Context(), projectID, uid, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": conversations, "total": total})
}

// DecodedConversationMessage represents a message with decoded payload
type DecodedConversationMessage struct {
	MessageID    int64                    `json:"message_id"`
	MessageIDStr string                   `json:"message_idstr"`
	MessageSeq   int64                    `json:"message_seq"`
	ClientMsgNo  string                   `json:"client_msg_no"`
	FromUID      string                   `json:"from_uid"`
	ChannelID    string                   `json:"channel_id"`
	ChannelType  int                      `json:"channel_type"`
	Timestamp    int64                    `json:"timestamp"`
	Payload      *wukongim.MessagePayload `json:"payload"`
}

// DecodedSyncConversation represents a conversation with decoded payloads
type DecodedSyncConversation struct {
	ChannelID   string                       `json:"channel_id"`
	ChannelType int                          `json:"channel_type"`
	UnreadCount int                          `json:"unread"`
	Timestamp   int64                        `json:"timestamp"`
	LastMsgSeq  int64                        `json:"last_msg_seq"`
	Version     int64                        `json:"version"`
	Recents     []DecodedConversationMessage `json:"recents"`
}

// decodeConversations decodes payloads in conversations
func decodeConversations(conversations []wukongim.SyncConversation) []DecodedSyncConversation {
	result := make([]DecodedSyncConversation, len(conversations))
	for i, conv := range conversations {
		result[i] = DecodedSyncConversation{
			ChannelID:   conv.ChannelID,
			ChannelType: conv.ChannelType,
			UnreadCount: conv.UnreadCount,
			Timestamp:   conv.Timestamp,
			LastMsgSeq:  conv.LastMsgSeq,
			Version:     conv.Version,
			Recents:     make([]DecodedConversationMessage, len(conv.Recents)),
		}
		for j, msg := range conv.Recents {
			result[i].Recents[j] = DecodedConversationMessage{
				MessageID:    msg.MessageID,
				MessageIDStr: msg.MessageIDStr,
				MessageSeq:   msg.MessageSeq,
				ClientMsgNo:  msg.ClientMsgNo,
				FromUID:      msg.FromUID,
				ChannelID:    msg.ChannelID,
				ChannelType:  msg.ChannelType,
				Timestamp:    msg.Timestamp,
				Payload:      wukongim.DecodeMessagePayload(msg.Payload),
			}
		}
	}
	return result
}

// GetMyConversations syncs the current staff's conversations from WuKongIM
func (h *ChatHandler) GetMyConversations(c *gin.Context) {
	// user_id is stored as uuid.UUID, not string
	userIDValue, _ := c.Get("user_id")
	userID := ""
	if uid, ok := userIDValue.(uuid.UUID); ok {
		userID = uid.String()
	}
	staffUID := userID + "-staff"

	var req struct {
		LastMsgSeqs map[string]int64 `json:"last_msg_seqs"`
		MsgCount    int              `json:"msg_count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// Use defaults if no body provided
		req.MsgCount = 20
	}
	if req.MsgCount == 0 {
		req.MsgCount = 20
	}

	conversations, err := h.svc.SyncMyConversations(c.Request.Context(), staffUID, req.LastMsgSeqs, req.MsgCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Decode payloads
	decoded := decodeConversations(conversations)

	c.JSON(http.StatusOK, gin.H{
		"conversations": decoded,
		"channels":      []interface{}{},
		"pagination": gin.H{
			"total":  len(decoded),
			"limit":  req.MsgCount,
			"offset": 0,
		},
	})
}

// GetWaitingConversations gets conversations for waiting visitors
func (h *ChatHandler) GetWaitingConversations(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	userID, _ := uuid.Parse(c.GetString("user_id"))
	staffUID := userID.String() + "-staff"

	msgCount, _ := strconv.Atoi(c.DefaultQuery("msg_count", "20"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	conversations, total, err := h.svc.GetWaitingConversations(c.Request.Context(), projectID, staffUID, msgCount, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Decode payloads
	decoded := decodeConversations(conversations)

	hasNext := (offset + limit) < total
	hasPrev := offset > 0

	c.JSON(http.StatusOK, gin.H{
		"conversations": decoded,
		"pagination": gin.H{
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"has_next": hasNext,
			"has_prev": hasPrev,
		},
	})
}

// GetAllConversations returns all conversations for the project
func (h *ChatHandler) GetAllConversations(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	conversations, total, err := h.svc.GetAllConversations(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hasNext := (offset + limit) < int(total)
	hasPrev := offset > 0

	c.JSON(http.StatusOK, gin.H{
		"conversations": conversations,
		"pagination": gin.H{
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"has_next": hasNext,
			"has_prev": hasPrev,
		},
	})
}

// GetConversationsByTagsRecent returns recent conversations by tags
func (h *ChatHandler) GetConversationsByTagsRecent(c *gin.Context) {
	projectID, _ := uuid.Parse(c.GetString("project_id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	conversations, total, err := h.svc.GetRecentConversations(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hasNext := (offset + limit) < int(total)
	hasPrev := offset > 0

	c.JSON(http.StatusOK, gin.H{
		"conversations": conversations,
		"pagination": gin.H{
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"has_next": hasNext,
			"has_prev": hasPrev,
		},
	})
}

// SetConversationUnread sets the unread count for a conversation
func (h *ChatHandler) SetConversationUnread(c *gin.Context) {
	// user_id is stored as uuid.UUID, not string
	userIDValue, _ := c.Get("user_id")
	userID := ""
	if uid, ok := userIDValue.(uuid.UUID); ok {
		userID = uid.String()
	}
	staffUID := userID + "-staff"

	var req struct {
		ChannelID   string `json:"channel_id"`
		ChannelType int    `json:"channel_type"`
		Unread      int    `json:"unread"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call WuKongIM to set unread count
	err := h.svc.SetConversationUnread(c.Request.Context(), staffUID, req.ChannelID, req.ChannelType, req.Unread)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ChatCompletionRequest represents visitor chat completion request
type ChatCompletionRequest struct {
	APIKey                       string `json:"api_key" binding:"required"`
	FromUID                      string `json:"from_uid"`
	Message                      string `json:"message" binding:"required"`
	ChannelID                    string `json:"channel_id"`
	ChannelType                  int    `json:"channel_type"`
	Stream                       bool   `json:"stream"`
	SystemMessage                string `json:"system_message"`
	ExpectedOutput               string `json:"expected_output"`
	ForwardUserMessageToWukongim bool   `json:"forward_user_message_to_wukongim"`
	WukongimOnly                 bool   `json:"wukongim_only"`
	TimeoutSeconds               int    `json:"timeout_seconds"`
}

// ChatCompletion handles visitor chat completion (public endpoint)
func (h *ChatHandler) ChatCompletion(c *gin.Context) {
	var req ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate platform API key
	platform, err := h.svc.ValidatePlatformAPIKey(c.Request.Context(), req.APIKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid platform API key"})
		return
	}

	// Get or create visitor
	visitorUID := req.FromUID
	if visitorUID == "" {
		visitorUID = uuid.New().String()
	}
	// Normalize: strip -vtr suffix for platform_open_id lookup
	platformOpenID := strings.TrimSuffix(visitorUID, "-vtr")

	visitor, err := h.svc.GetOrCreateVisitor(c.Request.Context(), platform, platformOpenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build channel ID
	channelID := req.ChannelID
	if channelID == "" {
		channelID = "cs_" + visitor.ID.String()
	}
	channelType := req.ChannelType
	if channelType == 0 {
		channelType = 251 // Customer service channel type
	}
	sessionID := channelID + "@" + strconv.Itoa(channelType)

	// Ensure visitor is subscribed to the channel
	visitorWkUID := visitor.ID.String() + "-vtr"
	h.svc.EnsureChannelSubscription(c.Request.Context(), channelID, channelType, visitorWkUID)

	// Forward user message to WuKongIM if requested
	if req.ForwardUserMessageToWukongim {
		h.svc.SendUserMessageToWukongim(c.Request.Context(), visitorWkUID, channelID, channelType, req.Message)
	}

	// Check if message is a manual service request (转人工)
	lowerMsg := strings.ToLower(req.Message)
	isManualServiceRequest := strings.Contains(req.Message, "转人工") ||
		strings.Contains(req.Message, "人工客服") ||
		strings.Contains(lowerMsg, "human") ||
		strings.Contains(lowerMsg, "transfer") ||
		strings.Contains(lowerMsg, "agent")

	if isManualServiceRequest {
		// Save visitor message first
		h.svc.SaveVisitorMessage(c.Request.Context(), platform.ProjectID, visitor.ID, channelID, req.Message)

		// Call internal AI events endpoint for transfer
		result, err := h.svc.TriggerManualServiceRequest(c.Request.Context(), platform.ProjectID, visitor.ID, req.Message)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success":    false,
				"event_type": "error",
				"message":    err.Error(),
				"visitor_id": visitor.ID.String(),
				"channel_id": channelID,
			})
			return
		}

		// Note: SendStaffAssignedMessage already sends the system message to WuKongIM
		// No need to send additional text message here

		c.JSON(http.StatusOK, result)
		return
	}

	// Save visitor message to database
	h.svc.SaveVisitorMessage(c.Request.Context(), platform.ProjectID, visitor.ID, channelID, req.Message)

	// Re-fetch visitor to get latest service_status (might have been updated by transfer)
	freshVisitor, _ := h.svc.GetVisitorByID(c.Request.Context(), platform.ProjectID, visitor.ID)
	if freshVisitor != nil {
		visitor = freshVisitor
	}

	// Check if visitor is in human service (skip AI if already transferred)
	if visitor.ServiceStatus == model.VisitorStatusActive {
		// Visitor is being served by human - refresh session TTL
		h.svc.RefreshHumanSession(c.Request.Context(), visitor.ID)

		// Don't send via WuKongIM here - Widget will send via WebSocket
		// Just return success so Widget knows to send via WebSocket
		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"event_type": "message_sent_human",
			"message":    "消息已发送给客服",
			"visitor_id": visitor.ID.String(),
			"channel_id": channelID,
		})
		return
	}

	// Call AI service with visitor context (non-streaming for now, aicenter streaming needs fix)
	resp, err := h.svc.CallAIServiceWithVisitor(c.Request.Context(), platform.ProjectID, &visitor.ID, req.Message, sessionID, req.SystemMessage, false)
	if err != nil {
		if strings.Contains(err.Error(), "ai_disabled") {
			c.JSON(http.StatusOK, gin.H{
				"success":    false,
				"event_type": "ai_disabled",
				"message":    "AI responses are disabled for this visitor/platform",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Save AI response to database and send to WuKongIM (SaveAIMessage handles both)
	h.svc.SaveAIMessage(c.Request.Context(), platform.ProjectID, channelID, resp.Content)

	// If wukongim_only, just return accepted status (message already sent via SaveAIMessage)
	if req.WukongimOnly {
		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"event_type": "accepted",
			"message":    "AI processing started",
			"visitor_id": visitor.ID.String(),
		})
		return
	}

	// Return success - AI response is sent via WuKongIM, no need to include content
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"event_type": "message_sent",
		"visitor_id": visitor.ID.String(),
		"channel_id": channelID,
	})
}

// streamChatCompletion handles streaming SSE response
func (h *ChatHandler) streamChatCompletion(c *gin.Context, platform *model.Platform, visitor *model.Visitor, channelID string, channelType int, sessionID string, req *ChatCompletionRequest) {
	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	// Call AI service with streaming
	streamChan, err := h.svc.CallAIServiceStream(c.Request.Context(), platform.ProjectID, req.Message, sessionID, req.SystemMessage)
	if err != nil {
		c.SSEvent("error", gin.H{"error": err.Error()})
		return
	}

	// Send initial event
	c.SSEvent("message", gin.H{
		"event_type": "start",
		"visitor_id": visitor.ID.String(),
		"channel_id": channelID,
	})
	c.Writer.Flush()

	// Stream chunks
	fullContent := ""
	for chunk := range streamChan {
		if chunk.Error != nil {
			c.SSEvent("error", gin.H{"error": chunk.Error.Error()})
			c.Writer.Flush()
			return
		}

		fullContent += chunk.Content
		c.SSEvent("message", gin.H{
			"event_type": "chunk",
			"content":    chunk.Content,
		})
		c.Writer.Flush()
	}

	// Send final event
	c.SSEvent("message", gin.H{
		"event_type": "end",
		"content":    fullContent,
		"visitor_id": visitor.ID.String(),
		"channel_id": channelID,
	})
	c.Writer.Flush()

	// If wukongim_only, send AI response to WuKongIM
	if req.WukongimOnly && fullContent != "" {
		staffUID := "ai-assistant"
		h.svc.SendAIResponseToWukongim(c.Request.Context(), staffUID, channelID, channelType, fullContent)
	}
}
