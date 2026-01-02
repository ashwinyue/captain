package db

import (
	"gorm.io/gorm"

	"github.com/tgo/captain/aicenter/internal/eino/memory"
	"github.com/tgo/captain/aicenter/internal/model"
)

// AutoMigrate runs GORM auto migration for all models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.LLMProvider{},
		&model.Team{},
		&model.Agent{},
		&model.AgentTool{},
		&model.AgentCollection{},
		&model.Tool{},
		&model.ProjectAIConfig{},
		&memory.ConversationMessage{}, // 会话记忆持久化
	)
}
