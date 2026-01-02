package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/pkg/wukongim"
	"github.com/tgo/captain/apiserver/internal/repository"
)

type ChatService struct {
	messageRepo      *repository.MessageRepository
	conversationRepo *repository.ConversationRepository
	queueRepo        *repository.QueueRepository
	wkClient         *wukongim.Client
	db               *gorm.DB
	aiCenterURL      string
	humanSessionSvc  *HumanSessionService
}

func NewChatService(messageRepo *repository.MessageRepository, conversationRepo *repository.ConversationRepository, queueRepo *repository.QueueRepository, wkClient *wukongim.Client) *ChatService {
	return &ChatService{messageRepo: messageRepo, conversationRepo: conversationRepo, queueRepo: queueRepo, wkClient: wkClient}
}

// NewChatServiceWithDB creates a ChatService with DB and AI center URL
func NewChatServiceWithDB(messageRepo *repository.MessageRepository, conversationRepo *repository.ConversationRepository, queueRepo *repository.QueueRepository, wkClient *wukongim.Client, db *gorm.DB, aiCenterURL string) *ChatService {
	svc := &ChatService{
		messageRepo:      messageRepo,
		conversationRepo: conversationRepo,
		queueRepo:        queueRepo,
		wkClient:         wkClient,
		db:               db,
		aiCenterURL:      aiCenterURL,
	}

	// Register ai-assistant user in WuKongIM on startup
	if wkClient != nil {
		go func() {
			ctx := context.Background()
			wkClient.RegisterOrLoginUser(ctx, "ai-assistant", "ai-assistant-token")
		}()
	}

	return svc
}

// SetHumanSessionService sets the human session service for timeout management
func (s *ChatService) SetHumanSessionService(svc *HumanSessionService) {
	s.humanSessionSvc = svc
}

// RefreshHumanSession refreshes the TTL for a visitor's human session
func (s *ChatService) RefreshHumanSession(ctx context.Context, visitorID uuid.UUID) {
	if s.humanSessionSvc != nil {
		s.humanSessionSvc.OnVisitorMessage(ctx, visitorID)
	}
}

type SendMessageRequest struct {
	ChannelID   string                 `json:"channel_id"`
	ChannelType int                    `json:"channel_type"`
	FromUID     string                 `json:"from_uid"`
	Content     string                 `json:"content"`
	MessageType model.MessageType      `json:"message_type"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

func (s *ChatService) SendMessage(ctx context.Context, projectID uuid.UUID, req *SendMessageRequest) (*model.Message, error) {
	payload := map[string]interface{}{
		"type":    req.MessageType,
		"content": req.Content,
	}
	if req.Extra != nil {
		payload["extra"] = req.Extra
	}
	payloadBytes, _ := json.Marshal(payload)

	wkResp, err := s.wkClient.SendMessage(ctx, &wukongim.SendMessageRequest{
		Header:      wukongim.MessageHeader{RedDot: 1},
		FromUID:     req.FromUID,
		ChannelID:   req.ChannelID,
		ChannelType: req.ChannelType,
		Payload:     payloadBytes,
	})
	if err != nil {
		return nil, err
	}

	msg := &model.Message{
		ProjectID:   projectID,
		MessageID:   fmt.Sprintf("%d", wkResp.MessageID),
		ChannelID:   req.ChannelID,
		FromUID:     req.FromUID,
		MessageType: req.MessageType,
		Content:     req.Content,
		Extra:       req.Extra,
		SentAt:      time.Now(),
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	s.updateConversation(ctx, projectID, req.ChannelID, req.FromUID, msg)
	return msg, nil
}

func (s *ChatService) GetMessages(ctx context.Context, projectID uuid.UUID, channelID string, limit, offset int) ([]model.Message, int64, error) {
	return s.messageRepo.FindByChannel(ctx, projectID, channelID, limit, offset)
}

// SaveVisitorMessage saves a visitor's message to the database (Widget sends to WuKongIM via WebSocket)
func (s *ChatService) SaveVisitorMessage(ctx context.Context, projectID uuid.UUID, visitorID uuid.UUID, channelID string, content string) error {
	fromUID := visitorID.String() + "-vtr"
	messageID := uuid.New().String()

	// Only save to database - Widget already sends to WuKongIM via WebSocket
	msg := &model.Message{
		ProjectID:   projectID,
		MessageID:   messageID,
		ChannelID:   channelID,
		FromUID:     fromUID,
		MessageType: model.MessageTypeText,
		Content:     content,
		SentAt:      time.Now(),
	}
	return s.messageRepo.Create(ctx, msg)
}

// SaveAIMessage saves an AI response message to the database and sends to WuKongIM
func (s *ChatService) SaveAIMessage(ctx context.Context, projectID uuid.UUID, channelID string, content string) error {
	messageID := uuid.New().String()

	// Send AI response to WuKongIM (visitor message is sent by Widget, but AI response needs to be sent by backend)
	if s.wkClient != nil {
		// Ensure channel exists before sending message
		s.wkClient.CreateOrUpdateChannel(ctx, channelID, 251, []string{"ai-assistant"})

		_, err := s.wkClient.SendTextMessage(ctx, &wukongim.SendTextMessageRequest{
			FromUID:     "ai-assistant",
			ChannelID:   channelID,
			ChannelType: 251, // Customer service channel
			Content:     content,
		})
		if err != nil {
			log.Printf("[SaveAIMessage] Failed to send to WuKongIM: %v", err)
		}
	}

	// Save to database
	msg := &model.Message{
		ProjectID:   projectID,
		MessageID:   messageID,
		ChannelID:   channelID,
		FromUID:     "ai-assistant",
		MessageType: model.MessageTypeText,
		Content:     content,
		SentAt:      time.Now(),
	}
	return s.messageRepo.Create(ctx, msg)
}

func (s *ChatService) RevokeMessage(ctx context.Context, projectID uuid.UUID, messageID string) error {
	return s.messageRepo.Revoke(ctx, projectID, messageID)
}

func (s *ChatService) updateConversation(ctx context.Context, projectID uuid.UUID, channelID, uid string, msg *model.Message) {
	now := time.Now()
	conv := &model.Conversation{
		ProjectID:     projectID,
		UID:           uid,
		ChannelID:     channelID,
		LastMessageID: msg.MessageID,
		LastMessage:   msg.Content,
		LastMessageAt: &now,
	}
	s.conversationRepo.Upsert(ctx, conv)
}

func (s *ChatService) GetConversations(ctx context.Context, projectID uuid.UUID, uid string, limit, offset int) ([]model.Conversation, int64, error) {
	return s.conversationRepo.FindByUID(ctx, projectID, uid, limit, offset)
}

func (s *ChatService) SyncMyConversations(ctx context.Context, staffUID string, lastMsgSeqs map[string]int64, msgCount int) ([]wukongim.SyncConversation, error) {
	return s.wkClient.SyncConversations(ctx, staffUID, lastMsgSeqs, msgCount)
}

// GetWaitingConversations gets conversations for waiting visitors
func (s *ChatService) GetWaitingConversations(ctx context.Context, projectID uuid.UUID, staffUID string, msgCount, limit, offset int) ([]wukongim.SyncConversation, int, error) {
	// 1. Get total count and waiting visitors
	total, err := s.queueRepo.CountWaiting(ctx, projectID)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []wukongim.SyncConversation{}, 0, nil
	}

	// 2. Get paginated waiting entries
	entries, err := s.queueRepo.FindWaitingPaginated(ctx, projectID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	if len(entries) == 0 {
		return []wukongim.SyncConversation{}, int(total), nil
	}

	// 3. Build channel list
	channels := make([]wukongim.ChannelInfo, 0, len(entries))
	for _, entry := range entries {
		channels = append(channels, wukongim.ChannelInfo{
			ChannelID:   entry.VisitorID.String(),
			ChannelType: 10, // Customer service channel type
		})
	}

	// 4. Sync conversations from WuKongIM
	conversations, err := s.wkClient.SyncConversationsByChannels(ctx, staffUID, channels, msgCount)
	if err != nil {
		return nil, 0, err
	}

	return conversations, int(total), nil
}

// GetAllConversations returns all conversations for the project
func (s *ChatService) GetAllConversations(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.Conversation, int64, error) {
	return s.conversationRepo.FindAll(ctx, projectID, limit, offset)
}

// GetRecentConversations returns recent conversations for the project
func (s *ChatService) GetRecentConversations(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.Conversation, int64, error) {
	return s.conversationRepo.FindRecent(ctx, projectID, limit, offset)
}

// SetConversationUnread sets the unread count for a conversation
func (s *ChatService) SetConversationUnread(ctx context.Context, uid, channelID string, channelType, unread int) error {
	return s.wkClient.SetConversationUnread(ctx, uid, channelID, channelType, unread)
}

// ValidatePlatformAPIKey validates a platform API key and returns the platform
func (s *ChatService) ValidatePlatformAPIKey(ctx context.Context, apiKey string) (*model.Platform, error) {
	var platform model.Platform
	err := s.db.WithContext(ctx).
		Where("api_key = ? AND is_active = true AND deleted_at IS NULL", apiKey).
		First(&platform).Error
	if err != nil {
		return nil, err
	}
	return &platform, nil
}

// GetVisitorByID gets a visitor by ID
func (s *ChatService) GetVisitorByID(ctx context.Context, projectID, visitorID uuid.UUID) (*model.Visitor, error) {
	var visitor model.Visitor
	err := s.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND deleted_at IS NULL", visitorID, projectID).
		First(&visitor).Error
	if err != nil {
		return nil, err
	}
	return &visitor, nil
}

// GetOrCreateVisitor gets or creates a visitor by platform open ID
func (s *ChatService) GetOrCreateVisitor(ctx context.Context, platform *model.Platform, platformOpenID string) (*model.Visitor, error) {
	var visitor model.Visitor
	err := s.db.WithContext(ctx).
		Where("project_id = ? AND platform_id = ? AND platform_open_id = ? AND deleted_at IS NULL",
			platform.ProjectID, platform.ID, platformOpenID).
		First(&visitor).Error

	if err == nil {
		// Update last visit time
		now := time.Now()
		visitor.LastVisitTime = &now
		s.db.Save(&visitor)
		return &visitor, nil
	}

	// Create new visitor
	now := time.Now()
	visitor = model.Visitor{
		ProjectID:      platform.ProjectID,
		PlatformID:     &platform.ID,
		PlatformOpenID: platformOpenID,
		Nickname:       platformOpenID,
		FirstSeenAt:    &now,
		LastSeenAt:     &now,
		LastVisitTime:  &now,
		VisitCount:     1,
		AIEnabled:      true,
	}
	if err := s.db.WithContext(ctx).Create(&visitor).Error; err != nil {
		return nil, err
	}
	return &visitor, nil
}

// SendUserMessageToWukongim sends user message to WuKongIM
func (s *ChatService) SendUserMessageToWukongim(ctx context.Context, fromUID, channelID string, channelType int, content string) {
	if s.wkClient == nil {
		return
	}
	s.wkClient.SendTextMessage(ctx, &wukongim.SendTextMessageRequest{
		FromUID:     fromUID,
		ChannelID:   channelID,
		ChannelType: channelType,
		Content:     content,
	})
}

// EnsureChannelSubscription ensures the user is subscribed to the channel
func (s *ChatService) EnsureChannelSubscription(ctx context.Context, channelID string, channelType int, userUID string) {
	if s.wkClient == nil {
		return
	}
	// Create/update channel with subscriber
	s.wkClient.CreateOrUpdateChannel(ctx, channelID, channelType, []string{userUID})
}

// ChatTransferResult represents the result of a transfer operation in chat
type ChatTransferResult struct {
	Success         bool       `json:"success"`
	AssignedStaffID *uuid.UUID `json:"assigned_staff_id,omitempty"`
	QueuePosition   *int       `json:"queue_position,omitempty"`
	Message         string     `json:"message"`
}

// TransferToStaff transfers a visitor to available staff
func (s *ChatService) TransferToStaff(ctx context.Context, projectID, visitorID uuid.UUID, channelID string, channelType int) (*ChatTransferResult, error) {
	// Find available staff
	var staffs []struct {
		ID uuid.UUID
	}
	if err := s.db.WithContext(ctx).Table("staff").
		Where("project_id = ? AND is_active = true AND service_paused = false AND deleted_at IS NULL", projectID).
		Find(&staffs).Error; err != nil || len(staffs) == 0 {
		// No available staff, add to queue
		queueEntry := &model.VisitorWaitingQueue{
			ProjectID:     projectID,
			VisitorID:     visitorID,
			ChannelID:     channelID,
			Status:        model.QueueStatusWaiting,
			WaitStartedAt: time.Now(),
		}
		if err := s.db.WithContext(ctx).Create(queueEntry).Error; err != nil {
			return &ChatTransferResult{Success: false, Message: "Failed to add to queue"}, nil
		}
		position := 1
		return &ChatTransferResult{Success: true, QueuePosition: &position, Message: "Added to queue"}, nil
	}

	// Assign first available staff
	assignedStaffID := staffs[0].ID

	// Add staff as subscriber to the channel
	if s.wkClient != nil {
		staffUID := assignedStaffID.String()
		s.wkClient.AddSubscribers(ctx, &wukongim.SubscribersRequest{
			ChannelID:   channelID,
			ChannelType: channelType,
			Subscribers: []string{staffUID},
		})
	}

	// Update visitor AI status
	s.db.WithContext(ctx).Table("visitors").
		Where("id = ?", visitorID).
		Update("ai_enabled", false)

	return &ChatTransferResult{
		Success:         true,
		AssignedStaffID: &assignedStaffID,
		Message:         "Transfer successful",
	}, nil
}

// AIServiceResponse represents AI service response
type AIServiceResponse struct {
	Content string `json:"content"`
}

// CallAIService calls the AI center service
func (s *ChatService) CallAIService(ctx context.Context, projectID uuid.UUID, message, sessionID, systemMessage string, stream bool) (*AIServiceResponse, error) {
	return s.CallAIServiceWithVisitor(ctx, projectID, nil, message, sessionID, systemMessage, stream)
}

// CallAIServiceWithVisitor calls the AI center service with visitor context for transfer to human
func (s *ChatService) CallAIServiceWithVisitor(ctx context.Context, projectID uuid.UUID, visitorID *uuid.UUID, message, sessionID, systemMessage string, stream bool) (*AIServiceResponse, error) {
	if s.aiCenterURL == "" {
		return nil, fmt.Errorf("AI center URL not configured")
	}

	// Get project's team from database
	teamID := ""
	var team struct {
		ID uuid.UUID
	}
	if err := s.db.WithContext(ctx).Table("ai_teams").
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		First(&team).Error; err == nil {
		teamID = team.ID.String()
	}

	// Prepare request
	reqBody := map[string]interface{}{
		"team_id":       teamID,
		"message":       message,
		"session_id":    sessionID,
		"stream":        false, // Always non-stream for now
		"enable_memory": true,
	}
	if systemMessage != "" {
		reqBody["system_message"] = systemMessage
	}
	if visitorID != nil {
		reqBody["visitor_id"] = visitorID.String()
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", s.aiCenterURL+"/api/v1/agents/run", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Project-ID", projectID.String())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	content := ""
	if c, ok := result["content"].(string); ok {
		content = c
	}

	return &AIServiceResponse{Content: content}, nil
}

// TriggerManualServiceRequest triggers a manual service request via internal API
// Aligns with original Python project's transfer_to_staff implementation
func (s *ChatService) TriggerManualServiceRequest(ctx context.Context, projectID, visitorID uuid.UUID, reason string) (map[string]interface{}, error) {
	channelID := "cs_" + visitorID.String()
	visitorUID := visitorID.String() + "-vtr"

	// 1. Check if visitor already in service (status is ACTIVE or QUEUED)
	var visitor model.Visitor
	if err := s.db.WithContext(ctx).Where("id = ?", visitorID).First(&visitor).Error; err == nil {
		if visitor.ServiceStatus == model.VisitorStatusActive {
			return map[string]interface{}{
				"success":    true,
				"event_type": "already_transferred",
				"message":    "您已在人工客服服务中",
				"content":    "您已在人工客服服务中",
				"visitor_id": visitorID.String(),
				"channel_id": channelID,
			}, nil
		}
		if visitor.ServiceStatus == model.VisitorStatusQueued {
			return map[string]interface{}{
				"success":    true,
				"event_type": "already_queued",
				"message":    "您已在等待队列中，请稍候",
				"content":    "您已在等待队列中，请稍候",
				"visitor_id": visitorID.String(),
				"channel_id": channelID,
			}, nil
		}
	}

	// 2. Find available staff
	var staffs []struct {
		ID       uuid.UUID
		Nickname string
		Username string
	}
	if err := s.db.WithContext(ctx).Table("staff").
		Select("id, nickname, username").
		Where("project_id = ? AND is_active = true AND service_paused = false AND deleted_at IS NULL", projectID).
		Find(&staffs).Error; err != nil || len(staffs) == 0 {
		// No available staff, add to queue
		now := time.Now()
		queueEntry := &model.VisitorWaitingQueue{
			ProjectID:     projectID,
			VisitorID:     visitorID,
			ChannelID:     channelID,
			Status:        model.QueueStatusWaiting,
			Source:        reason,
			WaitStartedAt: now,
		}
		if err := s.db.WithContext(ctx).Create(queueEntry).Error; err != nil {
			return nil, err
		}
		// Update visitor status to QUEUED
		s.db.WithContext(ctx).Model(&model.Visitor{}).
			Where("id = ?", visitorID).
			Updates(map[string]interface{}{
				"service_status": model.VisitorStatusQueued,
				"ai_enabled":     false,
			})

		// Send queued message to WuKongIM
		if s.wkClient != nil {
			s.wkClient.SendTextMessage(ctx, &wukongim.SendTextMessageRequest{
				FromUID:     "system",
				ChannelID:   channelID,
				ChannelType: 251,
				Content:     "当前没有可用客服，已加入等待队列，请稍候...",
			})
		}

		return map[string]interface{}{
			"success":        true,
			"event_type":     "queued",
			"message":        "当前没有可用客服，已加入等待队列",
			"content":        "当前没有可用客服，已加入等待队列",
			"visitor_id":     visitorID.String(),
			"channel_id":     channelID,
			"queue_position": 1,
		}, nil
	}

	// 3. Assign first available staff
	assignedStaff := staffs[0]
	staffUID := assignedStaff.ID.String() + "-staff"
	staffName := assignedStaff.Nickname
	if staffName == "" {
		staffName = assignedStaff.Username
	}
	if staffName == "" {
		staffName = "客服"
	}

	// 4. Create or get active session
	now := time.Now()
	session := &model.Session{
		ProjectID: projectID,
		ChannelID: channelID,
		VisitorID: visitorID,
		StaffID:   &assignedStaff.ID,
		Status:    model.SessionStatusActive,
		Source:    "manual_transfer",
		StartedAt: &now,
	}
	if err := s.db.WithContext(ctx).Create(session).Error; err != nil {
		log.Printf("[TriggerManualServiceRequest] Failed to create session: %v", err)
	}

	// 5. Add staff to ChannelMember table
	channelMember := &model.ChannelMember{
		ProjectID: projectID,
		ChannelID: channelID,
		UID:       staffUID,
	}
	s.db.WithContext(ctx).
		Where("channel_id = ? AND uid = ? AND deleted_at IS NULL", channelID, staffUID).
		FirstOrCreate(channelMember)

	// 6. Ensure channel exists and add subscribers to WuKongIM
	if s.wkClient != nil {
		// Create/update channel with visitor and staff as subscribers
		s.wkClient.CreateOrUpdateChannel(ctx, channelID, 251, []string{visitorUID, staffUID, "ai-assistant"})

		// Add staff as subscriber
		s.wkClient.AddSubscribers(ctx, &wukongim.SubscribersRequest{
			ChannelID:   channelID,
			ChannelType: 251,
			Subscribers: []string{staffUID},
		})

		// Note: Transfer success message is returned via API response for Widget to display locally
		// This avoids showing the message in the staff management system
	}

	// 7. Update visitor status to ACTIVE and assign staff
	s.db.WithContext(ctx).Model(&model.Visitor{}).
		Where("id = ?", visitorID).
		Updates(map[string]interface{}{
			"service_status":    model.VisitorStatusActive,
			"assigned_staff_id": assignedStaff.ID,
			"ai_enabled":        false,
		})

	// 8. Start human session timeout tracking in Redis
	if s.humanSessionSvc != nil {
		if err := s.humanSessionSvc.OnVisitorTransferToHuman(ctx, visitorID, assignedStaff.ID); err != nil {
			log.Printf("[TriggerManualServiceRequest] Failed to start human session tracking: %v", err)
		}
	}

	log.Printf("[TriggerManualServiceRequest] Transferred visitor %s to staff %s (%s)", visitorID, assignedStaff.ID, staffName)

	contentMsg := fmt.Sprintf("已为您转接人工客服 %s，请稍候...", staffName)
	return map[string]interface{}{
		"success":           true,
		"event_type":        "transfer_success",
		"message":           contentMsg,
		"content":           contentMsg,
		"visitor_id":        visitorID.String(),
		"channel_id":        channelID,
		"assigned_staff_id": assignedStaff.ID.String(),
	}, nil
}

// SendAIResponseToWukongim sends AI response to WuKongIM
func (s *ChatService) SendAIResponseToWukongim(ctx context.Context, fromUID, channelID string, channelType int, content string) {
	if s.wkClient == nil || content == "" {
		return
	}
	s.wkClient.SendTextMessage(ctx, &wukongim.SendTextMessageRequest{
		FromUID:     fromUID,
		ChannelID:   channelID,
		ChannelType: channelType,
		Content:     content,
	})
}

// CloseVisitorSession closes a visitor's service session
func (s *ChatService) CloseVisitorSession(ctx context.Context, projectID, visitorID uuid.UUID, closedByStaffID *uuid.UUID) error {
	channelID := "cs_" + visitorID.String()

	// Get visitor
	var visitor model.Visitor
	if err := s.db.WithContext(ctx).Where("id = ? AND project_id = ?", visitorID, projectID).First(&visitor).Error; err != nil {
		return fmt.Errorf("visitor not found: %w", err)
	}

	// Check if visitor is in active service
	if visitor.ServiceStatus != model.VisitorStatusActive {
		return fmt.Errorf("visitor is not in active service (current status: %s)", visitor.ServiceStatus)
	}

	// Update visitor status to CLOSED
	if err := s.db.WithContext(ctx).Model(&model.Visitor{}).
		Where("id = ?", visitorID).
		Updates(map[string]interface{}{
			"service_status":    model.VisitorStatusClosed,
			"assigned_staff_id": nil,
			"ai_enabled":        true, // Re-enable AI after session close
		}).Error; err != nil {
		return fmt.Errorf("failed to update visitor: %w", err)
	}

	// Remove staff from WuKongIM channel
	if s.wkClient != nil && visitor.AssignedStaffID != nil {
		staffUID := visitor.AssignedStaffID.String() + "-staff"
		s.wkClient.RemoveSubscribers(ctx, &wukongim.SubscribersRequest{
			ChannelID:   channelID,
			ChannelType: 251,
			Subscribers: []string{staffUID},
		})

		// Send session closed message
		s.wkClient.SendSessionClosedMessage(ctx, "system", channelID, 251, staffUID, "")
	}

	log.Printf("[CloseVisitorSession] Closed session for visitor %s", visitorID)
	return nil
}

// StreamChunk represents a chunk of streaming response
type StreamChunk struct {
	Content string
	Error   error
}

// CallAIServiceStream calls the AI center service with streaming
func (s *ChatService) CallAIServiceStream(ctx context.Context, projectID uuid.UUID, message, sessionID, systemMessage string) (<-chan StreamChunk, error) {
	if s.aiCenterURL == "" {
		return nil, fmt.Errorf("AI center URL not configured")
	}

	// Get project's team from database
	teamID := ""
	var team struct {
		ID uuid.UUID
	}
	if err := s.db.WithContext(ctx).Table("ai_teams").
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		First(&team).Error; err == nil {
		teamID = team.ID.String()
	}

	// Prepare request
	reqBody := map[string]interface{}{
		"team_id":       teamID,
		"message":       message,
		"session_id":    sessionID,
		"stream":        true,
		"enable_memory": true,
	}
	if systemMessage != "" {
		reqBody["system_message"] = systemMessage
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", s.aiCenterURL+"/api/v1/agents/run", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Project-ID", projectID.String())
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("AI service error: %s", string(body))
	}

	// Create channel for streaming
	chunkChan := make(chan StreamChunk)

	// Start goroutine to read SSE stream
	go func() {
		defer close(chunkChan)
		defer resp.Body.Close()

		reader := resp.Body
		buf := make([]byte, 4096)

		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err != io.EOF {
					chunkChan <- StreamChunk{Error: err}
				}
				return
			}

			// Parse SSE data
			data := string(buf[:n])
			lines := strings.Split(data, "\n")
			for _, line := range lines {
				// Handle data: lines
				if strings.HasPrefix(line, "data:") {
					jsonData := strings.TrimPrefix(line, "data:")
					jsonData = strings.TrimSpace(jsonData)
					if jsonData == "" || jsonData == "[DONE]" {
						continue
					}

					var event map[string]interface{}
					if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
						continue
					}

					// Check for done status
					if status, ok := event["status"].(string); ok && status == "done" {
						return
					}

					// Extract content from event (aicenter format: type="message", content="...")
					eventType, _ := event["type"].(string)
					if eventType == "message" {
						if content, ok := event["content"].(string); ok && content != "" {
							chunkChan <- StreamChunk{Content: content}
						}
					}

					// Also check direct content field
					if content, ok := event["content"].(string); ok && content != "" && eventType != "message" {
						chunkChan <- StreamChunk{Content: content}
					}

					// Check for chunk content
					if chunk, ok := event["chunk"].(string); ok && chunk != "" {
						chunkChan <- StreamChunk{Content: chunk}
					}
				}
			}
		}
	}()

	return chunkChan, nil
}
