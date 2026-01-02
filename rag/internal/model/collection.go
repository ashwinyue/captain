package model

import (
	"github.com/google/uuid"
)

type CollectionType string

const (
	CollectionTypeFile    CollectionType = "file"
	CollectionTypeWebsite CollectionType = "website"
	CollectionTypeQA      CollectionType = "qa"
)

type Collection struct {
	BaseModel
	ProjectID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	Name           string         `gorm:"size:255;not null" json:"name"`
	DisplayName    string         `gorm:"size:255" json:"display_name"`
	Description    string         `gorm:"type:text" json:"description"`
	CollectionType CollectionType `gorm:"size:50;not null" json:"collection_type"`
	EmbeddingModel string         `gorm:"size:100" json:"embedding_model"`
	ChunkSize      int            `gorm:"default:512" json:"chunk_size"`
	ChunkOverlap   int            `gorm:"default:50" json:"chunk_overlap"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	Tags           StringArray    `gorm:"type:jsonb" json:"tags"`
	Metadata       JSONMap        `gorm:"type:jsonb" json:"metadata"`

	// Stats (computed)
	DocumentCount int   `gorm:"-" json:"document_count,omitempty"`
	TotalSize     int64 `gorm:"-" json:"total_size,omitempty"`
}

func (Collection) TableName() string {
	return "rag_collections"
}
