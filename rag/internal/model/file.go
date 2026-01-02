package model

import (
	"time"

	"github.com/google/uuid"
)

type FileStatus string

const (
	FileStatusPending    FileStatus = "pending"
	FileStatusProcessing FileStatus = "processing"
	FileStatusCompleted  FileStatus = "completed"
	FileStatusFailed     FileStatus = "failed"
)

type File struct {
	BaseModel
	ProjectID    uuid.UUID   `gorm:"type:uuid;not null;index" json:"project_id"`
	CollectionID *uuid.UUID  `gorm:"type:uuid;index" json:"collection_id,omitempty"`
	FileName     string      `gorm:"size:500;not null" json:"file_name"`
	OriginalName string      `gorm:"size:500" json:"original_name"`
	ContentType  string      `gorm:"size:100" json:"content_type"`
	Size         int64       `gorm:"not null" json:"size"`
	StoragePath  string      `gorm:"size:1000" json:"storage_path"`
	Status       FileStatus  `gorm:"size:50;default:'pending'" json:"status"`
	ErrorMessage string      `gorm:"type:text" json:"error_message,omitempty"`
	UploadedBy   string      `gorm:"size:100" json:"uploaded_by"`
	ProcessedAt  *time.Time  `json:"processed_at,omitempty"`
	Tags         StringArray `gorm:"type:jsonb" json:"tags"`
	Metadata     JSONMap     `gorm:"type:jsonb" json:"metadata"`

	// Relations
	Collection *Collection `gorm:"foreignKey:CollectionID" json:"collection,omitempty"`
}

func (File) TableName() string {
	return "rag_files"
}
