package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
	"github.com/tgo/captain/apiserver/internal/pkg/wukongim"
	"github.com/tgo/captain/apiserver/internal/repository"
)

type VisitorService struct {
	repo     *repository.VisitorRepository
	wkClient *wukongim.Client
	db       *gorm.DB
}

func NewVisitorService(repo *repository.VisitorRepository, wkClient *wukongim.Client, db *gorm.DB) *VisitorService {
	return &VisitorService{repo: repo, wkClient: wkClient, db: db}
}

func (s *VisitorService) List(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]model.Visitor, int64, error) {
	visitors, total, err := s.repo.FindByProjectID(ctx, projectID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	// Populate ChannelID for each visitor
	for i := range visitors {
		visitors[i].ChannelID = BuildVisitorChannelID(visitors[i].ID)
	}
	return visitors, total, nil
}

func (s *VisitorService) GetByID(ctx context.Context, projectID, visitorID uuid.UUID) (*model.Visitor, error) {
	return s.repo.FindByIDAndProject(ctx, projectID, visitorID)
}

func (s *VisitorService) Create(ctx context.Context, visitor *model.Visitor) error {
	now := time.Now()
	visitor.FirstSeenAt = &now
	visitor.LastSeenAt = &now
	visitor.VisitCount = 1
	return s.repo.Create(ctx, visitor)
}

func (s *VisitorService) Update(ctx context.Context, visitor *model.Visitor) error {
	return s.repo.Update(ctx, visitor)
}

func (s *VisitorService) Delete(ctx context.Context, projectID, visitorID uuid.UUID) error {
	return s.repo.Delete(ctx, visitorID)
}

func (s *VisitorService) Block(ctx context.Context, projectID, visitorID uuid.UUID) error {
	return s.repo.UpdateBlocked(ctx, projectID, visitorID, true)
}

func (s *VisitorService) Unblock(ctx context.Context, projectID, visitorID uuid.UUID) error {
	return s.repo.UpdateBlocked(ctx, projectID, visitorID, false)
}

// VisitorRegisterRequest represents visitor registration request
type VisitorRegisterRequest struct {
	PlatformAPIKey   string                 `json:"platform_api_key"`
	PlatformOpenID   string                 `json:"platform_open_id,omitempty"`
	Name             string                 `json:"name,omitempty"`
	Nickname         string                 `json:"nickname,omitempty"`
	NicknameZh       string                 `json:"nickname_zh,omitempty"`
	AvatarURL        string                 `json:"avatar_url,omitempty"`
	PhoneNumber      string                 `json:"phone_number,omitempty"`
	Email            string                 `json:"email,omitempty"`
	Company          string                 `json:"company,omitempty"`
	JobTitle         string                 `json:"job_title,omitempty"`
	Source           string                 `json:"source,omitempty"`
	Note             string                 `json:"note,omitempty"`
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"`
	Timezone         string                 `json:"timezone,omitempty"`
	Language         string                 `json:"language,omitempty"`
	IPAddress        string                 `json:"ip_address,omitempty"`
}

// VisitorRegisterResponse represents visitor registration response
type VisitorRegisterResponse struct {
	*model.Visitor
	ChannelID   string `json:"channel_id"`
	ChannelType int    `json:"channel_type"`
	IMToken     string `json:"im_token"`
}

// BuildVisitorChannelID generates channel ID for visitor
func BuildVisitorChannelID(visitorID uuid.UUID) string {
	return "cs_" + visitorID.String()
}

// GetByChannelID gets visitor by channel ID
func (s *VisitorService) GetByChannelID(ctx context.Context, projectID uuid.UUID, channelID string) (*model.Visitor, error) {
	// Channel ID format: cs_{visitor_id}
	if !strings.HasPrefix(channelID, "cs_") {
		return nil, gorm.ErrRecordNotFound
	}
	visitorIDStr := strings.TrimPrefix(channelID, "cs_")
	visitorID, err := uuid.Parse(visitorIDStr)
	if err != nil {
		return nil, err
	}
	return s.repo.FindByIDAndProject(ctx, projectID, visitorID)
}

// SetAttributes sets visitor custom attributes
func (s *VisitorService) SetAttributes(ctx context.Context, projectID, visitorID uuid.UUID, attrs map[string]interface{}) (*model.Visitor, error) {
	visitor, err := s.repo.FindByIDAndProject(ctx, projectID, visitorID)
	if err != nil {
		return nil, err
	}

	// Update fields from attrs
	if name, ok := attrs["name"].(string); ok && name != "" {
		visitor.Name = name
	}
	if nickname, ok := attrs["nickname"].(string); ok && nickname != "" {
		visitor.Nickname = nickname
	}
	if nicknameZh, ok := attrs["nickname_zh"].(string); ok && nicknameZh != "" {
		visitor.NicknameZh = nicknameZh
	}
	if email, ok := attrs["email"].(string); ok && email != "" {
		visitor.Email = email
	}
	if phone, ok := attrs["phone_number"].(string); ok && phone != "" {
		visitor.PhoneNumber = phone
	}
	if company, ok := attrs["company"].(string); ok && company != "" {
		visitor.Company = company
	}
	if jobTitle, ok := attrs["job_title"].(string); ok && jobTitle != "" {
		visitor.JobTitle = jobTitle
	}
	if note, ok := attrs["note"].(string); ok {
		visitor.Note = note
	}
	if customAttrs, ok := attrs["custom_attributes"].(map[string]interface{}); ok {
		visitor.CustomAttributes = model.JSONMap(customAttrs)
	}

	if err := s.repo.Update(ctx, visitor); err != nil {
		return nil, err
	}
	return visitor, nil
}

// AcceptVisitorResponse represents accept visitor response
type AcceptVisitorResponse struct {
	*model.Visitor
	ChannelID   string `json:"channel_id"`
	ChannelType int    `json:"channel_type"`
	StaffUID    string `json:"staff_uid"`
}

// AcceptVisitor accepts a visitor and assigns to staff
func (s *VisitorService) AcceptVisitor(ctx context.Context, projectID, visitorID, staffID uuid.UUID) (*AcceptVisitorResponse, error) {
	visitor, err := s.repo.FindByIDAndProject(ctx, projectID, visitorID)
	if err != nil {
		return nil, err
	}

	// Check if visitor is already being served
	if visitor.ServiceStatus == model.VisitorStatusActive {
		return nil, fmt.Errorf("visitor is already being served (status is active)")
	}

	// Update visitor status
	visitor.ServiceStatus = model.VisitorStatusActive
	visitor.AIEnabled = false // Disable AI when human takes over
	visitor.AssignedStaffID = &staffID
	if err := s.db.WithContext(ctx).Save(visitor).Error; err != nil {
		return nil, err
	}

	channelID := BuildVisitorChannelID(visitorID)
	staffUID := staffID.String() + "-staff"

	// Add staff as subscriber to the channel (use channel type 251 for customer service)
	if s.wkClient != nil {
		s.wkClient.AddSubscribers(ctx, &wukongim.SubscribersRequest{
			ChannelID:   channelID,
			ChannelType: 251,
			Subscribers: []string{staffUID},
		})
	}

	return &AcceptVisitorResponse{
		Visitor:     visitor,
		ChannelID:   channelID,
		ChannelType: 251,
		StaffUID:    staffUID,
	}, nil
}

// SetAIEnabled enables or disables AI for visitor
func (s *VisitorService) SetAIEnabled(ctx context.Context, projectID, visitorID uuid.UUID, enabled bool) (*model.Visitor, error) {
	visitor, err := s.repo.FindByIDAndProject(ctx, projectID, visitorID)
	if err != nil {
		return nil, err
	}
	visitor.AIEnabled = enabled
	if err := s.repo.Update(ctx, visitor); err != nil {
		return nil, err
	}
	return visitor, nil
}

// SyncMessages syncs visitor messages
func (s *VisitorService) SyncMessages(ctx context.Context, platformAPIKey, channelID string, channelType int, startSeq, endSeq *int64, limit, pullMode int) (interface{}, error) {
	// Validate platform API key
	var platform model.Platform
	if err := s.db.WithContext(ctx).
		Where("api_key = ? AND is_active = true AND deleted_at IS NULL", platformAPIKey).
		First(&platform).Error; err != nil {
		return nil, err
	}

	// Sync messages from WuKongIM
	if s.wkClient != nil {
		var startSeqVal, endSeqVal int64
		if startSeq != nil {
			startSeqVal = *startSeq
		}
		if endSeq != nil {
			endSeqVal = *endSeq
		}
		// Extract visitor UID from channel ID (cs_xxx -> xxx-vtr)
		loginUID := ""
		if strings.HasPrefix(channelID, "cs_") {
			loginUID = strings.TrimPrefix(channelID, "cs_") + "-vtr"
		}
		return s.wkClient.SyncChannelMessages(ctx, loginUID, channelID, channelType, startSeqVal, endSeqVal, limit, pullMode)
	}
	return map[string]interface{}{"messages": []interface{}{}}, nil
}

// Register handles visitor registration via platform API key
func (s *VisitorService) Register(ctx context.Context, req *VisitorRegisterRequest, clientIP string) (*VisitorRegisterResponse, error) {
	// 1. Validate platform API key
	var platform model.Platform
	if err := s.db.WithContext(ctx).
		Where("api_key = ? AND is_active = true AND deleted_at IS NULL", req.PlatformAPIKey).
		First(&platform).Error; err != nil {
		return nil, err
	}

	// 2. Find existing or create new visitor
	platformOpenID := strings.TrimSpace(req.PlatformOpenID)
	var visitor *model.Visitor

	if platformOpenID != "" {
		visitor, _ = s.repo.FindByPlatformOpenID(ctx, platform.ProjectID, platform.ID, platformOpenID)
	}

	now := time.Now()

	// If nickname provided but nickname_zh not, set nickname_zh to nickname
	if req.Nickname != "" && req.NicknameZh == "" {
		req.NicknameZh = req.Nickname
	}

	// Use client IP if not provided
	ipAddress := req.IPAddress
	if ipAddress == "" {
		ipAddress = clientIP
	}

	if visitor == nil {
		visitor = &model.Visitor{
			ProjectID:      platform.ProjectID,
			PlatformID:     &platform.ID,
			PlatformOpenID: platformOpenID,
			Name:           req.Name,
			Nickname:       req.Nickname,
			NicknameZh:     req.NicknameZh,
			AvatarURL:      req.AvatarURL,
			PhoneNumber:    req.PhoneNumber,
			Email:          req.Email,
			Company:        req.Company,
			JobTitle:       req.JobTitle,
			Source:         req.Source,
			Note:           req.Note,
			Timezone:       req.Timezone,
			Language:       req.Language,
			IPAddress:      ipAddress,
			FirstSeenAt:    &now,
			LastSeenAt:     &now,
			LastVisitTime:  &now,
			VisitCount:     1,
			AIEnabled:      true,
		}
		if req.CustomAttributes != nil {
			visitor.CustomAttributes = model.JSONMap(req.CustomAttributes)
		}
		if err := s.repo.Create(ctx, visitor); err != nil {
			return nil, err
		}

		// Use visitor ID as platform_open_id if not provided
		if platformOpenID == "" {
			visitor.PlatformOpenID = visitor.ID.String()
			s.repo.Update(ctx, visitor)
		}
	} else {
		// Update existing visitor
		visitor.LastVisitTime = &now
		visitor.LastSeenAt = &now
		if req.Name != "" {
			visitor.Name = req.Name
		}
		if req.Nickname != "" {
			visitor.Nickname = req.Nickname
		}
		if req.NicknameZh != "" {
			visitor.NicknameZh = req.NicknameZh
		}
		if req.AvatarURL != "" {
			visitor.AvatarURL = req.AvatarURL
		}
		if req.PhoneNumber != "" {
			visitor.PhoneNumber = req.PhoneNumber
		}
		if req.Email != "" {
			visitor.Email = req.Email
		}
		if req.Company != "" {
			visitor.Company = req.Company
		}
		if req.JobTitle != "" {
			visitor.JobTitle = req.JobTitle
		}
		if req.Source != "" {
			visitor.Source = req.Source
		}
		if req.Note != "" {
			visitor.Note = req.Note
		}
		if req.Timezone != "" {
			visitor.Timezone = req.Timezone
		}
		if req.Language != "" {
			visitor.Language = req.Language
		}
		if ipAddress != "" {
			visitor.IPAddress = ipAddress
		}
		s.repo.Update(ctx, visitor)
	}

	// 3. Register visitor to WuKongIM
	imToken := uuid.New().String()
	visitorUID := visitor.ID.String() + "-vtr"
	if s.wkClient != nil {
		s.wkClient.RegisterOrLoginUser(ctx, visitorUID, imToken)
	}

	// 4. Create/update channel and ensure visitor is subscribed
	// Use channel type 251 for customer service (consistent with Widget)
	const customerServiceChannelType = 251
	channelID := BuildVisitorChannelID(visitor.ID)
	if s.wkClient != nil {
		// Create customer service channel with visitor as subscriber
		s.wkClient.CreateOrUpdateChannel(ctx, channelID, customerServiceChannelType, []string{visitorUID})
	}

	return &VisitorRegisterResponse{
		Visitor:     visitor,
		ChannelID:   channelID,
		ChannelType: customerServiceChannelType,
		IMToken:     imToken,
	}, nil
}
