package model

import (
	"time"

	"github.com/google/uuid"
)

// SystemSetup tracks system installation state
type SystemSetup struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	IsInstalled      bool       `gorm:"default:false" json:"is_installed"`
	AdminCreated     bool       `gorm:"default:false" json:"admin_created"`
	LLMConfigured    bool       `gorm:"default:false" json:"llm_configured"`
	SkipLLMConfig    bool       `gorm:"default:false" json:"skip_llm_config"`
	SetupVersion     string     `gorm:"size:20;default:'v1'" json:"setup_version"`
	SetupCompletedAt *time.Time `json:"setup_completed_at,omitempty"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SystemSetup) TableName() string {
	return "api_system_setup"
}

// RecalculateInstallFlags updates is_installed based on other flags
func (s *SystemSetup) RecalculateInstallFlags() bool {
	wasInstalled := s.IsInstalled
	s.IsInstalled = s.AdminCreated && (s.LLMConfigured || s.SkipLLMConfig)
	if s.IsInstalled && !wasInstalled && s.SetupCompletedAt == nil {
		now := time.Now()
		s.SetupCompletedAt = &now
	}
	return s.IsInstalled != wasInstalled
}
