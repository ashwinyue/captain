package model

import (
	"github.com/google/uuid"
)

type QAStatus string

const (
	QAStatusActive   QAStatus = "active"
	QAStatusInactive QAStatus = "inactive"
	QAStatusDraft    QAStatus = "draft"
)

type QAPair struct {
	BaseModel
	ProjectID    uuid.UUID   `gorm:"type:uuid;not null;index" json:"project_id"`
	CollectionID uuid.UUID   `gorm:"type:uuid;not null;index" json:"collection_id"`
	Question     string      `gorm:"type:text;not null" json:"question"`
	Answer       string      `gorm:"type:text;not null" json:"answer"`
	Category     string      `gorm:"size:100" json:"category"`
	Status       QAStatus    `gorm:"size:50;default:'active'" json:"status"`
	Priority     int         `gorm:"default:0" json:"priority"`
	VectorID     string      `gorm:"size:100" json:"vector_id"`
	Tags         StringArray `gorm:"type:jsonb" json:"tags"`
	Metadata     JSONMap     `gorm:"type:jsonb" json:"metadata"`

	// Relations
	Collection *Collection `gorm:"foreignKey:CollectionID" json:"collection,omitempty"`
}

func (QAPair) TableName() string {
	return "rag_qa_pairs"
}
