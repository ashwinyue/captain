package model

import (
	"github.com/google/uuid"
)

type Staff struct {
	BaseModel
	Username      string     `gorm:"size:100;uniqueIndex;not null" json:"username"`
	Email         string     `gorm:"size:255;uniqueIndex" json:"email,omitempty"`
	PasswordHash  string     `gorm:"size:255;not null" json:"-"`
	FullName      string     `gorm:"size:255" json:"full_name,omitempty"`
	Name          string     `gorm:"size:100" json:"name,omitempty"`
	Nickname      string     `gorm:"size:100" json:"nickname,omitempty"`
	Avatar        string     `gorm:"size:500" json:"avatar_url,omitempty"`
	Description   string     `gorm:"size:500" json:"description,omitempty"`
	Phone         string     `gorm:"size:50" json:"phone,omitempty"`
	Role          string     `gorm:"size:50;default:'user'" json:"role"`
	Status        string     `gorm:"size:20;default:'offline'" json:"status"`
	ProjectID     *uuid.UUID `gorm:"type:uuid;index" json:"project_id,omitempty"`
	IsActive      bool       `gorm:"default:true" json:"is_active"`
	IsSuperAdmin  bool       `gorm:"default:false" json:"is_super_admin"`
	ServicePaused bool       `gorm:"default:false" json:"service_paused"`
	LastLoginAt   *string    `json:"last_login_at,omitempty"`
	Extra         JSONMap    `gorm:"type:jsonb" json:"extra,omitempty"`
}

func (Staff) TableName() string {
	return "staff"
}
