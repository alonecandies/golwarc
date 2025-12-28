package models

import (
	"time"

	"gorm.io/gorm"
)

// Page represents a crawled web page
type Page struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	URL       string         `gorm:"uniqueIndex;not null;size:2048" json:"url"`
	Title     string         `gorm:"size:512" json:"title"`
	Content   string         `gorm:"type:longtext" json:"content"`
	Status    int            `gorm:"default:200" json:"status"`
	Domain    string         `gorm:"index;size:255" json:"domain"`
	HTML      string         `gorm:"type:longtext" json:"html,omitempty"`
	Headers   string         `gorm:"type:text" json:"headers,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Page model
func (Page) TableName() string {
	return "pages"
}
