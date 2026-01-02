package model

import (
	"time"

	"github.com/google/uuid"
)

type PageStatus string

const (
	PageStatusPending  PageStatus = "pending"
	PageStatusCrawling PageStatus = "crawling"
	PageStatusSuccess  PageStatus = "success"
	PageStatusFailed   PageStatus = "failed"
	PageStatusSkipped  PageStatus = "skipped"
)

type WebsitePage struct {
	BaseModel
	ProjectID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	CollectionID uuid.UUID  `gorm:"type:uuid;not null;index" json:"collection_id"`
	ParentPageID *uuid.UUID `gorm:"type:uuid;index" json:"parent_page_id,omitempty"`
	URL          string     `gorm:"size:2000;not null" json:"url"`
	Title        string     `gorm:"size:500" json:"title"`
	Description  string     `gorm:"type:text" json:"description"`
	Content      string     `gorm:"type:text" json:"content"`
	Status       PageStatus `gorm:"size:50;default:'pending'" json:"status"`
	Depth        int        `gorm:"default:0" json:"depth"`
	ErrorMessage string     `gorm:"type:text" json:"error_message,omitempty"`
	CrawledAt    *time.Time `json:"crawled_at,omitempty"`
	ContentHash  string     `gorm:"size:64" json:"content_hash"`
	Metadata     JSONMap    `gorm:"type:jsonb" json:"metadata"`

	// Relations
	Collection *Collection    `gorm:"foreignKey:CollectionID" json:"collection,omitempty"`
	ParentPage *WebsitePage   `gorm:"foreignKey:ParentPageID" json:"parent_page,omitempty"`
	ChildPages []*WebsitePage `gorm:"foreignKey:ParentPageID" json:"child_pages,omitempty"`
}

func (WebsitePage) TableName() string {
	return "rag_website_pages"
}
