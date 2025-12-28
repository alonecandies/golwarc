package models

import (
	"time"

	"gorm.io/gorm"
)

// Article represents a news article or blog post
type Article struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"not null;size:512" json:"title"`
	Author      string         `gorm:"index;size:255" json:"author"`
	Content     string         `gorm:"type:longtext" json:"content"`
	Summary     string         `gorm:"type:text" json:"summary"`
	PublishedAt *time.Time     `json:"published_at"`
	SourceURL   string         `gorm:"uniqueIndex;not null;size:2048" json:"source_url"`
	SourceName  string         `gorm:"index;size:255" json:"source_name"`
	Category    string         `gorm:"index;size:255" json:"category"`
	Tags        string         `gorm:"type:text" json:"tags"` // JSON array or comma-separated
	ImageURL    string         `gorm:"size:2048" json:"image_url"`
	Language    string         `gorm:"size:10;default:'en'" json:"language"`
	WordCount   int            `gorm:"default:0" json:"word_count"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Article model
func (Article) TableName() string {
	return "articles"
}
