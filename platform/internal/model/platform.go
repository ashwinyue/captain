package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// PlatformType represents the type of platform
type PlatformType string

const (
	PlatformTypeWeb       PlatformType = "web"
	PlatformTypeWechat    PlatformType = "wechat"
	PlatformTypeWhatsApp  PlatformType = "whatsapp"
	PlatformTypeTelegram  PlatformType = "telegram"
	PlatformTypeMessenger PlatformType = "messenger"
	PlatformTypeEmail     PlatformType = "email"
	PlatformTypeAPI       PlatformType = "api"
)

// Platform represents a communication platform
type Platform struct {
	BaseModel
	ProjectID   uuid.UUID    `gorm:"type:uuid;not null;index" json:"project_id"`
	Name        string       `gorm:"size:255;not null" json:"name"`
	Type        PlatformType `gorm:"size:50;not null" json:"type"`
	Description string       `gorm:"type:text" json:"description,omitempty"`
	Config      JSONMap      `gorm:"type:jsonb" json:"config,omitempty"`
	Credentials JSONMap      `gorm:"type:jsonb" json:"-"`
	IsEnabled   bool         `gorm:"default:true" json:"is_enabled"`
	WebhookURL  string       `gorm:"size:500" json:"webhook_url,omitempty"`
}

func (Platform) TableName() string {
	return "platforms"
}

// Onboarding represents user onboarding progress
type Onboarding struct {
	BaseModel
	ProjectID      uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"project_id"`
	CurrentStep    int        `gorm:"default:0" json:"current_step"`
	TotalSteps     int        `gorm:"default:5" json:"total_steps"`
	IsCompleted    bool       `gorm:"default:false" json:"is_completed"`
	CompletedSteps JSONMap    `gorm:"type:jsonb" json:"completed_steps,omitempty"`
	SkippedSteps   JSONMap    `gorm:"type:jsonb" json:"skipped_steps,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

func (Onboarding) TableName() string {
	return "onboardings"
}
