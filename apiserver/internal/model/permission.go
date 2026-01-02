package model

import (
	"github.com/google/uuid"
)

type RolePermission struct {
	BaseModel
	Role       string `gorm:"size:50;not null;index" json:"role"`
	Resource   string `gorm:"size:100;not null" json:"resource"`
	Action     string `gorm:"size:50;not null" json:"action"`
	Permission string `gorm:"size:150;not null;index" json:"permission"` // "resource:action"
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

type ProjectRolePermission struct {
	BaseModel
	ProjectID  uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	Role       string    `gorm:"size:50;not null;index" json:"role"`
	Resource   string    `gorm:"size:100;not null" json:"resource"`
	Action     string    `gorm:"size:50;not null" json:"action"`
	Permission string    `gorm:"size:150;not null;index" json:"permission"`
}

func (ProjectRolePermission) TableName() string {
	return "project_role_permissions"
}

const AdminRole = "admin"
