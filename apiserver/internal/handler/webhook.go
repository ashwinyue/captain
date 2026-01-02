package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WebhookHandler struct{}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{}
}

func (h *WebhookHandler) WuKongIMWebhook(c *gin.Context) {
	event := c.Query("event")
	log.Printf("WuKongIM webhook received: event=%s", event)

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "invalid payload"})
		return
	}

	switch event {
	case "user.onlinestatus":
		// Handle online status events
		go h.processOnlineStatus(body)
	case "msg.notify":
		// Handle message notifications
		go h.processMsgNotify(body)
	default:
		log.Printf("Unhandled WuKongIM webhook event: %s", event)
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

func (h *WebhookHandler) processOnlineStatus(payload map[string]interface{}) {
	// Extract user info from payload
	uid, _ := payload["uid"].(string)
	deviceFlag, _ := payload["device_flag"].(float64)
	online, _ := payload["online"].(float64)

	log.Printf("User online status changed: uid=%s, device=%d, online=%v", uid, int(deviceFlag), online == 1)

	// Update user's online status in database
	// This would typically update the visitor's last_seen_at and online status
	// For now, just log the event
	if online == 1 {
		log.Printf("User %s came online on device %d", uid, int(deviceFlag))
	} else {
		log.Printf("User %s went offline from device %d", uid, int(deviceFlag))
	}
}

func (h *WebhookHandler) processMsgNotify(payload map[string]interface{}) {
	// Extract message info from payload
	fromUID, _ := payload["from_uid"].(string)
	channelID, _ := payload["channel_id"].(string)
	channelType, _ := payload["channel_type"].(float64)
	msgPayload, _ := payload["payload"].(map[string]interface{})

	log.Printf("Message notification: from=%s, channel=%s, type=%d", fromUID, channelID, int(channelType))

	// Extract message content
	content := ""
	if msgPayload != nil {
		if c, ok := msgPayload["content"].(string); ok {
			content = c
		}
	}

	// Process the message based on channel type
	// Channel types: 1=person, 2=group
	switch int(channelType) {
	case 1:
		// Personal message
		log.Printf("Personal message from %s: %s", fromUID, content)
		// Here you would typically:
		// 1. Find or create a conversation
		// 2. Store the message
		// 3. Trigger AI response if configured
	case 2:
		// Group message
		log.Printf("Group message in %s from %s: %s", channelID, fromUID, content)
	default:
		log.Printf("Unknown channel type %d", int(channelType))
	}
}
