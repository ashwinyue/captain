package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/platform/internal/model"
)

type CallbackHandler struct {
	db *gorm.DB
}

func NewCallbackHandler(db *gorm.DB) *CallbackHandler {
	return &CallbackHandler{db: db}
}

func computeMsgSignature(token, timestamp, nonce string, msg ...string) string {
	parts := []string{token, timestamp, nonce}
	parts = append(parts, msg...)
	sort.Strings(parts)
	h := sha1.New()
	h.Write([]byte(strings.Join(parts, "")))
	return hex.EncodeToString(h.Sum(nil))
}

// WeComVerify handles WeCom URL verification (GET)
func (h *CallbackHandler) WeComVerify(c *gin.Context) {
	platformAPIKey := c.Param("platform_api_key")

	var platform model.Platform
	if err := h.db.Where("api_key = ? AND is_enabled = ?", platformAPIKey, true).First(&platform).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid platform_api_key"})
		return
	}

	config := platform.Config
	token, _ := config["token"].(string)
	msgSignature := c.Query("msg_signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echostr")

	if token == "" || msgSignature == "" || echostr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required params"})
		return
	}

	expected := computeMsgSignature(token, timestamp, nonce, echostr)
	if expected != msgSignature {
		c.JSON(http.StatusForbidden, gin.H{"error": "signature mismatch"})
		return
	}

	// For encrypted mode, decrypt echostr; for plain mode, return as-is
	// TODO: Implement AES decryption for encrypted mode
	c.String(http.StatusOK, echostr)
}

// WeComCallback handles WeCom webhook POST callback
func (h *CallbackHandler) WeComCallback(c *gin.Context) {
	platformAPIKey := c.Param("platform_api_key")

	var platform model.Platform
	if err := h.db.Where("api_key = ? AND is_enabled = ?", platformAPIKey, true).First(&platform).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid platform_api_key"})
		return
	}

	config := platform.Config
	token, _ := config["token"].(string)
	msgSignature := c.Query("msg_signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Parse XML
	type XMLMessage struct {
		XMLName      xml.Name `xml:"xml"`
		ToUserName   string   `xml:"ToUserName"`
		FromUserName string   `xml:"FromUserName"`
		CreateTime   int64    `xml:"CreateTime"`
		MsgType      string   `xml:"MsgType"`
		Content      string   `xml:"Content"`
		MsgId        string   `xml:"MsgId"`
		Encrypt      string   `xml:"Encrypt"`
	}

	var xmlMsg XMLMessage
	if err := xml.Unmarshal(body, &xmlMsg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid XML payload"})
		return
	}

	// Verify signature
	if xmlMsg.Encrypt != "" {
		expected := computeMsgSignature(token, timestamp, nonce, xmlMsg.Encrypt)
		if expected != msgSignature {
			c.JSON(http.StatusForbidden, gin.H{"error": "signature mismatch"})
			return
		}
		// TODO: Decrypt encrypted message
	} else {
		expected := computeMsgSignature(token, timestamp, nonce)
		if expected != msgSignature {
			c.JSON(http.StatusForbidden, gin.H{"error": "signature mismatch"})
			return
		}
	}

	// Store in inbox
	now := time.Now()
	receivedAt := time.Unix(xmlMsg.CreateTime, 0)
	inbox := model.WeComInbox{
		PlatformID:  platform.ID,
		MessageID:   xmlMsg.MsgId,
		SourceType:  "wecom",
		FromUser:    xmlMsg.FromUserName,
		MsgType:     xmlMsg.MsgType,
		Content:     xmlMsg.Content,
		RawPayload:  model.JSONMap{"raw_xml": string(body)},
		Status:      model.InboxStatusPending,
		ReceivedAt:  &receivedAt,
		ProcessedAt: nil,
	}
	inbox.ID = uuid.New()
	inbox.CreatedAt = now
	inbox.UpdatedAt = now

	if err := h.db.Create(&inbox).Error; err != nil {
		// Duplicate message, ignore
		if strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusOK, gin.H{"ok": true})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// FeishuCallback handles Feishu webhook callback
func (h *CallbackHandler) FeishuCallback(c *gin.Context) {
	platformAPIKey := c.Param("platform_api_key")

	var platform model.Platform
	if err := h.db.Where("api_key = ? AND is_enabled = ?", platformAPIKey, true).First(&platform).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid platform_api_key"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Parse JSON payload
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		// Re-read body for binding
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}

	// Handle URL verification challenge
	if challenge, ok := payload["challenge"].(string); ok {
		c.JSON(http.StatusOK, gin.H{"challenge": challenge})
		return
	}

	// Extract event data
	header, _ := payload["header"].(map[string]interface{})
	event, _ := payload["event"].(map[string]interface{})

	if header == nil || event == nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	eventID, _ := header["event_id"].(string)
	message, _ := event["message"].(map[string]interface{})
	sender, _ := event["sender"].(map[string]interface{})

	if message == nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	messageID, _ := message["message_id"].(string)
	chatID, _ := message["chat_id"].(string)
	chatType, _ := message["chat_type"].(string)
	msgType, _ := message["message_type"].(string)
	content, _ := message["content"].(string)

	senderID := ""
	senderType := ""
	if sender != nil {
		senderID, _ = sender["sender_id"].(map[string]interface{})["open_id"].(string)
		senderType, _ = sender["sender_type"].(string)
	}

	// Store in inbox
	now := time.Now()
	inbox := model.FeishuInbox{
		PlatformID: platform.ID,
		MessageID:  messageID,
		ChatID:     chatID,
		ChatType:   chatType,
		SenderID:   senderID,
		SenderType: senderType,
		MsgType:    msgType,
		Content:    content,
		RawPayload: model.JSONMap{"event_id": eventID, "raw": string(body)},
		Status:     model.InboxStatusPending,
		ReceivedAt: &now,
	}
	inbox.ID = uuid.New()
	inbox.CreatedAt = now
	inbox.UpdatedAt = now

	if err := h.db.Create(&inbox).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusOK, gin.H{"ok": true})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DingTalkCallback handles DingTalk webhook callback
func (h *CallbackHandler) DingTalkCallback(c *gin.Context) {
	platformAPIKey := c.Param("platform_api_key")

	var platform model.Platform
	if err := h.db.Where("api_key = ? AND is_enabled = ?", platformAPIKey, true).First(&platform).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid platform_api_key"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}

	// Extract message data
	msgID, _ := payload["msgId"].(string)
	conversationID, _ := payload["conversationId"].(string)
	conversationType, _ := payload["conversationType"].(string)
	senderID, _ := payload["senderId"].(string)
	senderNick, _ := payload["senderNick"].(string)
	msgType, _ := payload["msgtype"].(string)

	content := ""
	if text, ok := payload["text"].(map[string]interface{}); ok {
		content, _ = text["content"].(string)
	}

	// Store in inbox
	now := time.Now()
	inbox := model.DingTalkInbox{
		PlatformID:       platform.ID,
		MessageID:        msgID,
		ConversationID:   conversationID,
		ConversationType: conversationType,
		SenderID:         senderID,
		SenderNick:       senderNick,
		MsgType:          msgType,
		Content:          content,
		RawPayload:       model.JSONMap{"raw": string(body)},
		Status:           model.InboxStatusPending,
		ReceivedAt:       &now,
	}
	inbox.ID = uuid.New()
	inbox.CreatedAt = now
	inbox.UpdatedAt = now

	if err := h.db.Create(&inbox).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusOK, gin.H{"ok": true})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// WuKongIMCallback handles WuKongIM webhook callback
func (h *CallbackHandler) WuKongIMCallback(c *gin.Context) {
	platformAPIKey := c.Param("platform_api_key")

	var platform model.Platform
	if err := h.db.Where("api_key = ? AND is_enabled = ?", platformAPIKey, true).First(&platform).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid platform_api_key"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}

	// Extract message data
	msgID, _ := payload["message_id"].(string)
	channelID, _ := payload["channel_id"].(string)
	channelType := int(payload["channel_type"].(float64))
	fromUID, _ := payload["from_uid"].(string)
	msgType := int(payload["msg_type"].(float64))
	content, _ := payload["content"].(string)

	// Store in inbox
	now := time.Now()
	inbox := model.WuKongIMInbox{
		PlatformID:  platform.ID,
		MessageID:   msgID,
		ChannelID:   channelID,
		ChannelType: channelType,
		FromUID:     fromUID,
		MsgType:     msgType,
		Content:     content,
		RawPayload:  model.JSONMap{"raw": string(body)},
		Status:      model.InboxStatusPending,
		ReceivedAt:  &now,
	}
	inbox.ID = uuid.New()
	inbox.CreatedAt = now
	inbox.UpdatedAt = now

	if err := h.db.Create(&inbox).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusOK, gin.H{"ok": true})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
