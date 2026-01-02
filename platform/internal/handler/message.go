package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/tgo/captain/platform/internal/model"
)

type MessageHandler struct {
	db *gorm.DB
}

func NewMessageHandler(db *gorm.DB) *MessageHandler {
	return &MessageHandler{db: db}
}

type IngestRequest struct {
	PlatformAPIKey string                 `json:"platform_api_key" binding:"required"`
	SourceType     string                 `json:"source_type"`
	MessageID      string                 `json:"message_id"`
	FromUser       string                 `json:"from_user"`
	MsgType        string                 `json:"msg_type"`
	Content        string                 `json:"content"`
	RawPayload     map[string]interface{} `json:"raw_payload"`
}

// Ingest receives a normalized message and processes it
func (h *MessageHandler) Ingest(c *gin.Context) {
	var req IngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Lookup platform
	var platform model.Platform
	if err := h.db.Where("api_key = ? AND is_enabled = ?", req.PlatformAPIKey, true).First(&platform).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "platform not found"})
		return
	}

	// TODO: Process message through normalizer and dispatcher
	// For now, just acknowledge receipt
	_ = platform

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type SendMessageRequest struct {
	PlatformAPIKey string                 `json:"platform_api_key" binding:"required"`
	FromUID        string                 `json:"from_uid"`
	ChannelID      string                 `json:"channel_id" binding:"required"`
	ChannelType    int                    `json:"channel_type"`
	Payload        map[string]interface{} `json:"payload" binding:"required"`
	ClientMsgNo    string                 `json:"client_msg_no"`
}

// SendMessage sends a message to a third-party platform
func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Lookup platform
	var platform model.Platform
	if err := h.db.Where("api_key = ? AND is_enabled = ?", req.PlatformAPIKey, true).First(&platform).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "platform not found"})
		return
	}

	// TODO: Implement message sending based on platform type
	// For now, return a placeholder response
	_ = platform

	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"client_msg_no": req.ClientMsgNo,
		"message":       "Message sending not yet implemented",
	})
}
