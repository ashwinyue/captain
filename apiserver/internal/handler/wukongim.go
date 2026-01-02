package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/tgo/captain/apiserver/internal/pkg/wukongim"
	"github.com/tgo/captain/apiserver/internal/service"
)

type WuKongIMHandler struct {
	svc *service.WuKongIMService
}

func NewWuKongIMHandler(svc *service.WuKongIMService) *WuKongIMHandler {
	return &WuKongIMHandler{svc: svc}
}

func (h *WuKongIMHandler) GetRoute(c *gin.Context) {
	uid := c.Query("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uid parameter is required"})
		return
	}

	route, err := h.svc.GetRoute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, route)
}

func (h *WuKongIMHandler) SyncChannelMessages(c *gin.Context) {
	var req service.ChannelMessageSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert channel_id format: if it ends with -vtr, convert to cs_ format
	if strings.HasSuffix(req.ChannelID, "-vtr") {
		// Extract visitor ID and convert to cs_ format
		visitorID := strings.TrimSuffix(req.ChannelID, "-vtr")
		req.ChannelID = "cs_" + visitorID
	}

	result, err := h.svc.SyncChannelMessages(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	// Decode payload for each message
	if result != nil && result.Messages != nil {
		for i, msg := range result.Messages {
			if msgMap, ok := msg.(map[string]interface{}); ok {
				if payload, ok := msgMap["payload"].(string); ok {
					decoded := wukongim.DecodeMessagePayload(payload)
					if decoded != nil {
						msgMap["payload"] = decoded
					}
				}
				result.Messages[i] = msgMap
			}
		}
	}

	c.JSON(http.StatusOK, result)
}

func (h *WuKongIMHandler) GetChannelInfo(c *gin.Context) {
	channelID := c.Query("channel_id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id parameter is required"})
		return
	}

	channelTypeStr := c.Query("channel_type")
	channelType := 1 // default
	if channelTypeStr != "" {
		var err error
		channelType, err = strconv.Atoi(channelTypeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_type"})
			return
		}
	}

	// Parse channel_id to determine entity type and return channel info
	result := h.svc.GetChannelInfoByID(c.Request.Context(), channelID, channelType)
	c.JSON(http.StatusOK, result)
}
