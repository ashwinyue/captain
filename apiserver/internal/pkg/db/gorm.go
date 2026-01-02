package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/model"
)

func NewGormDB(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.SystemSetup{},
		&model.Staff{},
		&model.Project{},
		&model.RolePermission{},
		&model.ProjectRolePermission{},
		&model.Visitor{},
		&model.Tag{},
		&model.VisitorTag{},
		&model.VisitorWaitingQueue{},
		&model.AssignmentRule{},
		&model.Channel{},
		&model.ChannelMember{},
		&model.Session{},
		&model.Conversation{},
		&model.Message{},
		&model.Platform{},
		&model.AIProvider{},
	)
}
