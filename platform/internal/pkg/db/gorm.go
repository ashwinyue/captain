package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/tgo/captain/platform/internal/model"
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
		&model.Platform{},
		&model.Onboarding{},
	)
}
