package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/pkg/wukongim"
	"gorm.io/gorm"
)

type WuKongIMService struct {
	client  *wukongim.Client
	baseURL string
	db      *gorm.DB
}

func NewWuKongIMService(client *wukongim.Client, baseURL string, db *gorm.DB) *WuKongIMService {
	return &WuKongIMService{client: client, baseURL: baseURL, db: db}
}

type RouteResponse struct {
	TCPAddr string `json:"tcp_addr"`
	WSAddr  string `json:"ws_addr"`
}

func (s *WuKongIMService) GetRoute(ctx context.Context, uid string) (*RouteResponse, error) {
	url := fmt.Sprintf("%s/route?uid=%s", s.baseURL, uid)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get route: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("WuKongIM route request failed with status %d", resp.StatusCode)
	}

	var result RouteResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode route response: %w", err)
	}

	return &result, nil
}

type ChannelMessageSyncRequest struct {
	LoginUID        string `json:"login_uid"`
	ChannelID       string `json:"channel_id"`
	ChannelType     int    `json:"channel_type"`
	StartMessageSeq int64  `json:"start_message_seq"`
	EndMessageSeq   int64  `json:"end_message_seq"`
	Limit           int    `json:"limit"`
	PullMode        int    `json:"pull_mode"`
}

func (s *WuKongIMService) SyncChannelMessages(ctx context.Context, req *ChannelMessageSyncRequest) (*wukongim.ChannelMessageSyncResponse, error) {
	return s.client.SyncChannelMessages(ctx, req.LoginUID, req.ChannelID, req.ChannelType, req.StartMessageSeq, req.EndMessageSeq, req.Limit, req.PullMode)
}

func (s *WuKongIMService) GetChannelInfo(ctx context.Context, channelID string, channelType int) (*wukongim.ChannelInfoResponse, error) {
	return s.client.GetChannelInfo(ctx, channelID, channelType)
}

// ChannelInfoResult represents the channel info response for frontend
type ChannelInfoResult struct {
	Name        string                 `json:"name"`
	Avatar      string                 `json:"avatar"`
	ChannelID   string                 `json:"channel_id"`
	ChannelType int                    `json:"channel_type"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// GetChannelInfoByID parses channel_id to determine entity type and returns channel info
func (s *WuKongIMService) GetChannelInfoByID(ctx context.Context, channelID string, channelType int) *ChannelInfoResult {
	result := &ChannelInfoResult{
		ChannelID:   channelID,
		ChannelType: channelType,
		Extra:       make(map[string]interface{}),
	}

	// Parse channel_id suffix to determine type
	switch {
	case strings.HasSuffix(channelID, "-agent"):
		result.Name = "AI Agent"
		result.Avatar = ""
		result.Extra["type"] = "agent"
	case strings.HasSuffix(channelID, "-team"):
		result.Name = "AI Team"
		result.Avatar = ""
		result.Extra["type"] = "team"
	case strings.HasSuffix(channelID, "-staff"):
		result.Name = "Staff"
		result.Avatar = ""
		result.Extra["type"] = "staff"
	case strings.HasSuffix(channelID, "-vtr"):
		result.Name = "Visitor"
		result.Avatar = ""
		result.Extra["type"] = "visitor"
	case strings.HasPrefix(channelID, "cs_"):
		// Customer service channel - fetch visitor info from database
		visitorIDStr := strings.TrimPrefix(channelID, "cs_")
		visitorID, err := uuid.Parse(visitorIDStr)
		if err == nil && s.db != nil {
			var visitor model.Visitor
			if err := s.db.WithContext(ctx).Where("id = ?", visitorID).First(&visitor).Error; err == nil {
				result.Name = visitor.Nickname
				if visitor.NicknameZh != "" {
					result.Name = visitor.NicknameZh
				}
				result.Avatar = visitor.AvatarURL
				result.Extra["type"] = "visitor"
				result.Extra["id"] = visitor.ID.String()
				result.Extra["service_status"] = visitor.ServiceStatus
				result.Extra["ai_enabled"] = visitor.AIEnabled
				result.Extra["ai_disabled"] = !visitor.AIEnabled
				if visitor.AssignedStaffID != nil {
					result.Extra["assigned_staff_id"] = visitor.AssignedStaffID.String()
				}
				result.Extra["is_online"] = visitor.IsOnline
				return result
			}
		}
		result.Name = "Visitor"
		result.Extra["type"] = "visitor"
	default:
		result.Name = "Channel"
		result.Avatar = ""
		result.Extra["type"] = "unknown"
	}

	return result
}
