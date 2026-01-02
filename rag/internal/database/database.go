package database

import (
	"github.com/tgo/captain/rag/internal/config"
	"github.com/tgo/captain/rag/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Warn
	if cfg.IsDevelopment() {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Collection{},
		&model.File{},
		&model.Document{},
		&model.WebsitePage{},
		&model.QAPair{},
		&model.EmbeddingConfig{},
	)
}
