package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UUIDArray is a custom type for UUID arrays in PostgreSQL
type UUIDArray []uuid.UUID

func (a UUIDArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *UUIDArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, a)
}

// PlatformType represents platform type enumeration
type PlatformType string

const (
	PlatformTypeWebsite     PlatformType = "website"
	PlatformTypeWechat      PlatformType = "wechat"
	PlatformTypeWhatsapp    PlatformType = "whatsapp"
	PlatformTypeTelegram    PlatformType = "telegram"
	PlatformTypeEmail       PlatformType = "email"
	PlatformTypeSMS         PlatformType = "sms"
	PlatformTypeFacebook    PlatformType = "facebook"
	PlatformTypeInstagram   PlatformType = "instagram"
	PlatformTypeTwitter     PlatformType = "twitter"
	PlatformTypeLinkedin    PlatformType = "linkedin"
	PlatformTypeDiscord     PlatformType = "discord"
	PlatformTypeSlack       PlatformType = "slack"
	PlatformTypeTeams       PlatformType = "teams"
	PlatformTypePhone       PlatformType = "phone"
	PlatformTypeDouyin      PlatformType = "douyin"
	PlatformTypeTiktok      PlatformType = "tiktok"
	PlatformTypeCustom      PlatformType = "custom"
	PlatformTypeWecom       PlatformType = "wecom"
	PlatformTypeWecomBot    PlatformType = "wecom_bot"
	PlatformTypeFeishuBot   PlatformType = "feishu_bot"
	PlatformTypeDingtalkBot PlatformType = "dingtalk_bot"
)

// PlatformSyncStatus represents sync status
type PlatformSyncStatus string

const (
	PlatformSyncPending PlatformSyncStatus = "pending"
	PlatformSyncSynced  PlatformSyncStatus = "synced"
	PlatformSyncFailed  PlatformSyncStatus = "failed"
)

// PlatformAIMode represents AI mode
type PlatformAIMode string

const (
	PlatformAIModeAuto   PlatformAIMode = "auto"   // AI handles all messages
	PlatformAIModeAssist PlatformAIMode = "assist" // Human first, AI fallback
	PlatformAIModeOff    PlatformAIMode = "off"    // AI disabled
)

// PlatformTypeDefinition represents platform type metadata
type PlatformTypeDefinition struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Type        string    `gorm:"size:50;uniqueIndex;not null" json:"type"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	NameEN      *string   `gorm:"size:100;column:name_en" json:"name_en,omitempty"`
	IsSupported bool      `gorm:"default:true" json:"is_supported"`
	Icon        *string   `gorm:"type:text" json:"icon,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (PlatformTypeDefinition) TableName() string {
	return "api_platform_types"
}

// Platform represents a communication platform
type Platform struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	Name      *string   `gorm:"size:100" json:"name,omitempty"`
	Type      string    `gorm:"size:20;not null" json:"type"`
	APIKey    *string   `gorm:"size:255" json:"api_key,omitempty"`
	Config    JSONMap   `gorm:"type:jsonb" json:"config,omitempty"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`

	// AI configuration
	AgentIDs            UUIDArray `gorm:"type:jsonb;column:agent_ids" json:"agent_ids,omitempty"`
	AIMode              *string   `gorm:"size:20;default:'auto'" json:"ai_mode,omitempty"`
	FallbackToAITimeout *int      `gorm:"default:0" json:"fallback_to_ai_timeout,omitempty"`

	// Website usage tracking
	IsUsed           bool    `gorm:"default:false" json:"is_used"`
	UsedWebsiteURL   *string `gorm:"size:1024" json:"used_website_url,omitempty"`
	UsedWebsiteTitle *string `gorm:"size:255" json:"used_website_title,omitempty"`

	// Logo storage
	LogoPath *string `gorm:"size:512" json:"logo_path,omitempty"`

	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Sync tracking
	SyncStatus     string     `gorm:"size:20;default:'pending'" json:"sync_status"`
	LastSyncedAt   *time.Time `json:"last_synced_at,omitempty"`
	SyncError      *string    `gorm:"type:text" json:"sync_error,omitempty"`
	SyncRetryCount int        `gorm:"default:0" json:"sync_retry_count"`

	// Relations
	Project      *Project                `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	PlatformType *PlatformTypeDefinition `gorm:"foreignKey:Type;references:Type" json:"platform_type,omitempty"`
}

func (Platform) TableName() string {
	return "api_platforms"
}

// IsDeleted checks if the platform is soft deleted
func (p *Platform) IsDeleted() bool {
	return p.DeletedAt.Valid
}

// GetIcon returns SVG icon markup for the platform type
func (p *Platform) GetIcon() *string {
	if p.PlatformType != nil {
		return p.PlatformType.Icon
	}
	return nil
}

// GetIsSupported returns whether this platform type is currently supported
func (p *Platform) GetIsSupported() *bool {
	if p.PlatformType != nil {
		return &p.PlatformType.IsSupported
	}
	return nil
}

// GetNameEN returns English name of the platform type
func (p *Platform) GetNameEN() *string {
	if p.PlatformType != nil {
		return p.PlatformType.NameEN
	}
	return nil
}
