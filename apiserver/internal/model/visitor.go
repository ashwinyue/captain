package model

import (
	"time"

	"github.com/google/uuid"
)

// VisitorServiceStatus represents the service status of a visitor
type VisitorServiceStatus string

const (
	// VisitorStatusNew - Visitor just created, no service requested
	VisitorStatusNew VisitorServiceStatus = "new"
	// VisitorStatusQueued - In waiting queue for human service
	VisitorStatusQueued VisitorServiceStatus = "queued"
	// VisitorStatusActive - Currently being served by staff
	VisitorStatusActive VisitorServiceStatus = "active"
	// VisitorStatusClosed - Service session closed
	VisitorStatusClosed VisitorServiceStatus = "closed"
)

type Visitor struct {
	BaseModel
	ProjectID        uuid.UUID            `gorm:"type:uuid;not null;index" json:"project_id"`
	PlatformID       *uuid.UUID           `gorm:"type:uuid;index" json:"platform_id,omitempty"`
	PlatformOpenID   string               `gorm:"size:255;index" json:"platform_open_id,omitempty"`
	ChannelID        string               `gorm:"-" json:"channel_id,omitempty"` // Computed field, not stored in DB
	ExternalID       string               `gorm:"size:255;index" json:"external_id,omitempty"`
	Name             string               `gorm:"size:255" json:"name"`
	Nickname         string               `gorm:"size:255" json:"nickname,omitempty"`
	NicknameZh       string               `gorm:"size:255;column:nickname_zh" json:"nickname_zh,omitempty"`
	Email            string               `gorm:"size:255;index" json:"email,omitempty"`
	Phone            string               `gorm:"size:50" json:"phone,omitempty"`
	PhoneNumber      string               `gorm:"size:50" json:"phone_number,omitempty"`
	Avatar           string               `gorm:"size:500" json:"avatar,omitempty"`
	AvatarURL        string               `gorm:"size:500" json:"avatar_url,omitempty"`
	Source           string               `gorm:"size:100" json:"source,omitempty"`
	Country          string               `gorm:"size:100" json:"country,omitempty"`
	City             string               `gorm:"size:100" json:"city,omitempty"`
	IPAddress        string               `gorm:"size:45" json:"ip_address,omitempty"`
	Language         string               `gorm:"size:20" json:"language,omitempty"`
	Timezone         string               `gorm:"size:50" json:"timezone,omitempty"`
	Browser          string               `gorm:"size:100" json:"browser,omitempty"`
	OS               string               `gorm:"size:100" json:"os,omitempty"`
	Device           string               `gorm:"size:100" json:"device,omitempty"`
	Company          string               `gorm:"size:255" json:"company,omitempty"`
	JobTitle         string               `gorm:"size:255" json:"job_title,omitempty"`
	Note             string               `gorm:"type:text" json:"note,omitempty"`
	FirstSeenAt      *time.Time           `json:"first_seen_at,omitempty"`
	LastSeenAt       *time.Time           `json:"last_seen_at,omitempty"`
	LastVisitTime    *time.Time           `json:"last_visit_time,omitempty"`
	VisitCount       int                  `gorm:"default:0" json:"visit_count"`
	IsBlocked        bool                 `gorm:"default:false" json:"is_blocked"`
	IsOnline         bool                 `gorm:"default:false" json:"is_online"`
	AIEnabled        bool                 `gorm:"default:true" json:"ai_enabled"`
	ServiceStatus    VisitorServiceStatus `gorm:"size:20;default:'new'" json:"service_status"`
	AssignedStaffID  *uuid.UUID           `gorm:"type:uuid;index" json:"assigned_staff_id,omitempty"`
	Tags             JSONMap              `gorm:"type:jsonb" json:"tags,omitempty"`
	CustomData       JSONMap              `gorm:"type:jsonb" json:"custom_data,omitempty"`
	CustomAttributes JSONMap              `gorm:"type:jsonb" json:"custom_attributes,omitempty"`
}

// IsUnassigned checks if visitor can be assigned to staff
func (v *Visitor) IsUnassigned() bool {
	return v.ServiceStatus == VisitorStatusNew || v.ServiceStatus == VisitorStatusClosed
}

func (Visitor) TableName() string {
	return "visitors"
}

type Tag struct {
	BaseModel
	ProjectID   uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Color       string    `gorm:"size:20" json:"color,omitempty"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Category    string    `gorm:"size:50" json:"category,omitempty"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
}

func (Tag) TableName() string {
	return "tags"
}

type VisitorTag struct {
	BaseModel
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	VisitorID uuid.UUID `gorm:"type:uuid;not null;index" json:"visitor_id"`
	TagID     uuid.UUID `gorm:"type:uuid;not null;index" json:"tag_id"`
}

func (VisitorTag) TableName() string {
	return "visitor_tags"
}

type QueueStatus string

const (
	QueueStatusWaiting  QueueStatus = "waiting"
	QueueStatusAssigned QueueStatus = "assigned"
	QueueStatusTimeout  QueueStatus = "timeout"
	QueueStatusLeft     QueueStatus = "left"
)

type VisitorWaitingQueue struct {
	BaseModel
	ProjectID     uuid.UUID   `gorm:"type:uuid;not null;index" json:"project_id"`
	VisitorID     uuid.UUID   `gorm:"type:uuid;not null;index" json:"visitor_id"`
	ChannelID     string      `gorm:"size:100;index" json:"channel_id"`
	Status        QueueStatus `gorm:"size:50;default:'waiting'" json:"status"`
	Priority      int         `gorm:"default:0" json:"priority"`
	Source        string      `gorm:"size:100" json:"source,omitempty"`
	AssignedTo    *uuid.UUID  `gorm:"type:uuid" json:"assigned_to,omitempty"`
	AssignedAt    *time.Time  `json:"assigned_at,omitempty"`
	WaitStartedAt time.Time   `gorm:"not null" json:"wait_started_at"`
	Extra         JSONMap     `gorm:"type:jsonb" json:"extra,omitempty"`
	Visitor       *Visitor    `gorm:"foreignKey:VisitorID" json:"visitor,omitempty"`
}

func (VisitorWaitingQueue) TableName() string {
	return "visitor_waiting_queue"
}

type AssignmentRule struct {
	BaseModel
	ProjectID            uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"project_id"`
	Name                 string     `gorm:"size:255;not null" json:"name"`
	Description          string     `gorm:"type:text" json:"description,omitempty"`
	Priority             int        `gorm:"default:0" json:"priority"`
	IsEnabled            bool       `gorm:"default:true" json:"is_enabled"`
	AIProviderID         *uuid.UUID `gorm:"type:uuid" json:"ai_provider_id,omitempty"`
	Model                string     `gorm:"size:100" json:"model,omitempty"`
	Prompt               string     `gorm:"type:text" json:"prompt,omitempty"`
	LLMAssignmentEnabled bool       `gorm:"default:false" json:"llm_assignment_enabled"`
	Timezone             string     `gorm:"size:50;default:'UTC'" json:"timezone"`
	ServiceWeekdays      []int      `gorm:"type:jsonb;serializer:json" json:"service_weekdays"`
	ServiceStartTime     string     `gorm:"size:10;default:'09:00'" json:"service_start_time"`
	ServiceEndTime       string     `gorm:"size:10;default:'18:00'" json:"service_end_time"`
	MaxConcurrentChats   int        `gorm:"default:5" json:"max_concurrent_chats"`
	AutoCloseHours       int        `gorm:"default:24" json:"auto_close_hours"`
	Conditions           JSONMap    `gorm:"type:jsonb" json:"conditions,omitempty"`
	Actions              JSONMap    `gorm:"type:jsonb" json:"actions,omitempty"`
}

func (AssignmentRule) TableName() string {
	return "visitor_assignment_rules"
}
