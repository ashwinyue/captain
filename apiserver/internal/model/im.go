package model

import (
	"time"

	"github.com/google/uuid"
)

type ChannelType int

const (
	ChannelTypePersonal ChannelType = 1
	ChannelTypeGroup    ChannelType = 2
)

type Channel struct {
	BaseModel
	ProjectID   uuid.UUID   `gorm:"type:uuid;not null;index" json:"project_id"`
	ChannelID   string      `gorm:"size:100;not null;uniqueIndex" json:"channel_id"`
	ChannelType ChannelType `gorm:"not null" json:"channel_type"`
	Name        string      `gorm:"size:255" json:"name"`
	Avatar      string      `gorm:"size:500" json:"avatar"`
	Extra       JSONMap     `gorm:"type:jsonb" json:"extra,omitempty"`
	IsDisabled  bool        `gorm:"default:false" json:"is_disabled"`
}

func (Channel) TableName() string {
	return "im_channels"
}

type ChannelMember struct {
	BaseModel
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	ChannelID string    `gorm:"size:100;not null;index" json:"channel_id"`
	UID       string    `gorm:"size:100;not null;index" json:"uid"`
	Role      int       `gorm:"default:0" json:"role"`
	IsMuted   bool      `gorm:"default:false" json:"is_muted"`
	Extra     JSONMap   `gorm:"type:jsonb" json:"extra,omitempty"`
}

func (ChannelMember) TableName() string {
	return "im_channel_members"
}

type SessionStatus string

const (
	SessionStatusActive   SessionStatus = "active"
	SessionStatusClosed   SessionStatus = "closed"
	SessionStatusPending  SessionStatus = "pending"
	SessionStatusTransfer SessionStatus = "transfer"
)

type Session struct {
	BaseModel
	ProjectID       uuid.UUID     `gorm:"type:uuid;not null;index" json:"project_id"`
	ChannelID       string        `gorm:"size:100;not null;index" json:"channel_id"`
	VisitorID       uuid.UUID     `gorm:"type:uuid;index" json:"visitor_id"`
	StaffID         *uuid.UUID    `gorm:"type:uuid;index" json:"staff_id,omitempty"`
	Status          SessionStatus `gorm:"size:50;default:'pending'" json:"status"`
	Source          string        `gorm:"size:100" json:"source"`
	StartedAt       *time.Time    `json:"started_at,omitempty"`
	EndedAt         *time.Time    `json:"ended_at,omitempty"`
	ClosedByStaffID *uuid.UUID    `gorm:"type:uuid" json:"closed_by_staff_id,omitempty"`
	DurationSeconds int           `gorm:"default:0" json:"duration_seconds"`
	LastMessageSeq  int64         `gorm:"default:0" json:"last_message_seq"`
	LastMessageAt   *time.Time    `json:"last_message_at,omitempty"`
	MessageCount    int           `gorm:"default:0" json:"message_count"`
	Extra           JSONMap       `gorm:"type:jsonb" json:"extra,omitempty"`
}

// Close closes the session and calculates duration
func (s *Session) Close(closedByStaffID *uuid.UUID) {
	now := time.Now()
	s.Status = SessionStatusClosed
	s.EndedAt = &now
	s.ClosedByStaffID = closedByStaffID
	if s.StartedAt != nil {
		s.DurationSeconds = int(now.Sub(*s.StartedAt).Seconds())
	}
}

func (Session) TableName() string {
	return "im_sessions"
}

type Conversation struct {
	BaseModel
	ProjectID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	UID           string     `gorm:"size:100;not null;index" json:"uid"`
	ChannelID     string     `gorm:"size:100;not null;index" json:"channel_id"`
	LastMessageID string     `gorm:"size:100" json:"last_message_id"`
	LastMessage   string     `gorm:"type:text" json:"last_message"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	UnreadCount   int        `gorm:"default:0" json:"unread_count"`
	IsPinned      bool       `gorm:"default:false" json:"is_pinned"`
	IsMuted       bool       `gorm:"default:false" json:"is_muted"`
	Extra         JSONMap    `gorm:"type:jsonb" json:"extra,omitempty"`
}

func (Conversation) TableName() string {
	return "im_conversations"
}

type MessageType int

const (
	MessageTypeText   MessageType = 1
	MessageTypeImage  MessageType = 2
	MessageTypeVoice  MessageType = 3
	MessageTypeVideo  MessageType = 4
	MessageTypeFile   MessageType = 5
	MessageTypeSystem MessageType = 99
)

type Message struct {
	BaseModel
	ProjectID   uuid.UUID   `gorm:"type:uuid;not null;index" json:"project_id"`
	MessageID   string      `gorm:"size:100;not null;uniqueIndex" json:"message_id"`
	ChannelID   string      `gorm:"size:100;not null;index" json:"channel_id"`
	FromUID     string      `gorm:"size:100;not null;index" json:"from_uid"`
	MessageType MessageType `gorm:"not null" json:"message_type"`
	Content     string      `gorm:"type:text" json:"content"`
	Extra       JSONMap     `gorm:"type:jsonb" json:"extra,omitempty"`
	SentAt      time.Time   `gorm:"not null" json:"sent_at"`
	IsRevoked   bool        `gorm:"default:false" json:"is_revoked"`
}

func (Message) TableName() string {
	return "im_messages"
}
