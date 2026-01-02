package model

import (
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type Document struct {
	BaseModel
	ProjectID    uuid.UUID       `gorm:"type:uuid;not null;index" json:"project_id"`
	CollectionID uuid.UUID       `gorm:"type:uuid;not null;index" json:"collection_id"`
	FileID       *uuid.UUID      `gorm:"type:uuid;index" json:"file_id,omitempty"`
	PageID       *uuid.UUID      `gorm:"type:uuid;index" json:"page_id,omitempty"`
	Content      string          `gorm:"type:text;not null" json:"content"`
	ChunkIndex   int             `gorm:"default:0" json:"chunk_index"`
	TokenCount   int             `gorm:"default:0" json:"token_count"`
	VectorID     string          `gorm:"size:100" json:"vector_id"`
	Embedding    pgvector.Vector `gorm:"type:vector(1536)" json:"-"` // Standard embedding dimension (1536)
	Metadata     JSONMap         `gorm:"type:jsonb" json:"metadata"`

	// Relations
	Collection *Collection `gorm:"foreignKey:CollectionID" json:"collection,omitempty"`
	File       *File       `gorm:"foreignKey:FileID" json:"file,omitempty"`
}

func (Document) TableName() string {
	return "rag_documents"
}
