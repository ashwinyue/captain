package model

import (
	"time"

	"github.com/google/uuid"
)

type InboxStatus string

const (
	InboxStatusPending   InboxStatus = "pending"
	InboxStatusProcessed InboxStatus = "processed"
	InboxStatusFailed    InboxStatus = "failed"
)

// WeComInbox stores incoming WeCom messages
type WeComInbox struct {
	BaseModel
	PlatformID      uuid.UUID   `gorm:"type:uuid;not null;index" json:"platform_id"`
	MessageID       string      `gorm:"size:100;uniqueIndex" json:"message_id"`
	SourceType      string      `gorm:"size:50" json:"source_type"`
	FromUser        string      `gorm:"size:255" json:"from_user"`
	OpenKfID        string      `gorm:"size:100" json:"open_kfid,omitempty"`
	MsgType         string      `gorm:"size:50" json:"msg_type"`
	Content         string      `gorm:"type:text" json:"content"`
	IsFromColleague bool        `gorm:"default:false" json:"is_from_colleague"`
	RawPayload      JSONMap     `gorm:"type:jsonb" json:"raw_payload,omitempty"`
	Status          InboxStatus `gorm:"size:50;default:'pending'" json:"status"`
	ReceivedAt      *time.Time  `json:"received_at,omitempty"`
	ProcessedAt     *time.Time  `json:"processed_at,omitempty"`
	ErrorMessage    string      `gorm:"type:text" json:"error_message,omitempty"`
}

func (WeComInbox) TableName() string {
	return "wecom_inbox"
}

// FeishuInbox stores incoming Feishu messages
type FeishuInbox struct {
	BaseModel
	PlatformID   uuid.UUID   `gorm:"type:uuid;not null;index" json:"platform_id"`
	MessageID    string      `gorm:"size:100;uniqueIndex" json:"message_id"`
	ChatID       string      `gorm:"size:100" json:"chat_id"`
	ChatType     string      `gorm:"size:50" json:"chat_type"`
	SenderID     string      `gorm:"size:100" json:"sender_id"`
	SenderType   string      `gorm:"size:50" json:"sender_type"`
	MsgType      string      `gorm:"size:50" json:"msg_type"`
	Content      string      `gorm:"type:text" json:"content"`
	RawPayload   JSONMap     `gorm:"type:jsonb" json:"raw_payload,omitempty"`
	Status       InboxStatus `gorm:"size:50;default:'pending'" json:"status"`
	ReceivedAt   *time.Time  `json:"received_at,omitempty"`
	ProcessedAt  *time.Time  `json:"processed_at,omitempty"`
	ErrorMessage string      `gorm:"type:text" json:"error_message,omitempty"`
}

func (FeishuInbox) TableName() string {
	return "feishu_inbox"
}

// DingTalkInbox stores incoming DingTalk messages
type DingTalkInbox struct {
	BaseModel
	PlatformID       uuid.UUID   `gorm:"type:uuid;not null;index" json:"platform_id"`
	MessageID        string      `gorm:"size:100;uniqueIndex" json:"message_id"`
	ConversationID   string      `gorm:"size:100" json:"conversation_id"`
	ConversationType string      `gorm:"size:50" json:"conversation_type"`
	SenderID         string      `gorm:"size:100" json:"sender_id"`
	SenderNick       string      `gorm:"size:255" json:"sender_nick"`
	MsgType          string      `gorm:"size:50" json:"msg_type"`
	Content          string      `gorm:"type:text" json:"content"`
	RawPayload       JSONMap     `gorm:"type:jsonb" json:"raw_payload,omitempty"`
	Status           InboxStatus `gorm:"size:50;default:'pending'" json:"status"`
	ReceivedAt       *time.Time  `json:"received_at,omitempty"`
	ProcessedAt      *time.Time  `json:"processed_at,omitempty"`
	ErrorMessage     string      `gorm:"type:text" json:"error_message,omitempty"`
}

func (DingTalkInbox) TableName() string {
	return "dingtalk_inbox"
}

// WuKongIMInbox stores incoming WuKongIM messages
type WuKongIMInbox struct {
	BaseModel
	PlatformID   uuid.UUID   `gorm:"type:uuid;not null;index" json:"platform_id"`
	MessageID    string      `gorm:"size:100;uniqueIndex" json:"message_id"`
	ChannelID    string      `gorm:"size:100" json:"channel_id"`
	ChannelType  int         `gorm:"default:0" json:"channel_type"`
	FromUID      string      `gorm:"size:100" json:"from_uid"`
	MsgType      int         `gorm:"default:1" json:"msg_type"`
	Content      string      `gorm:"type:text" json:"content"`
	RawPayload   JSONMap     `gorm:"type:jsonb" json:"raw_payload,omitempty"`
	Status       InboxStatus `gorm:"size:50;default:'pending'" json:"status"`
	ReceivedAt   *time.Time  `json:"received_at,omitempty"`
	ProcessedAt  *time.Time  `json:"processed_at,omitempty"`
	ErrorMessage string      `gorm:"type:text" json:"error_message,omitempty"`
}

func (WuKongIMInbox) TableName() string {
	return "wukongim_inbox"
}

// EmailInbox stores incoming email messages
type EmailInbox struct {
	BaseModel
	PlatformID   uuid.UUID   `gorm:"type:uuid;not null;index" json:"platform_id"`
	MessageID    string      `gorm:"size:255;uniqueIndex" json:"message_id"`
	FromAddress  string      `gorm:"size:255" json:"from_address"`
	ToAddress    string      `gorm:"size:255" json:"to_address"`
	Subject      string      `gorm:"size:500" json:"subject"`
	BodyText     string      `gorm:"type:text" json:"body_text"`
	BodyHTML     string      `gorm:"type:text" json:"body_html"`
	RawPayload   JSONMap     `gorm:"type:jsonb" json:"raw_payload,omitempty"`
	Status       InboxStatus `gorm:"size:50;default:'pending'" json:"status"`
	ReceivedAt   *time.Time  `json:"received_at,omitempty"`
	ProcessedAt  *time.Time  `json:"processed_at,omitempty"`
	ErrorMessage string      `gorm:"type:text" json:"error_message,omitempty"`
}

func (EmailInbox) TableName() string {
	return "email_inbox"
}
