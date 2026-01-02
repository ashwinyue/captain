package model

type Project struct {
	BaseModel
	Name        string  `gorm:"size:255;not null" json:"name"`
	Slug        string  `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Description string  `gorm:"type:text" json:"description,omitempty"`
	APIKey      string  `gorm:"size:255;uniqueIndex" json:"api_key,omitempty"`
	Logo        string  `gorm:"size:500" json:"logo,omitempty"`
	IsActive    bool    `gorm:"default:true" json:"is_active"`
	Settings    JSONMap `gorm:"type:jsonb" json:"settings,omitempty"`
}

func (Project) TableName() string {
	return "projects"
}
