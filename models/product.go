package models

import (
	"time"

	"gorm.io/gorm"
)

// Product represents a product scraped from e-commerce sites
type Product struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null;size:512" json:"name"`
	Price       float64        `gorm:"type:decimal(10,2)" json:"price"`
	Currency    string         `gorm:"size:10;default:'USD'" json:"currency"`
	Description string         `gorm:"type:text" json:"description"`
	ImageURL    string         `gorm:"size:2048" json:"image_url"`
	SourceURL   string         `gorm:"uniqueIndex;not null;size:2048" json:"source_url"`
	Category    string         `gorm:"index;size:255" json:"category"`
	Brand       string         `gorm:"index;size:255" json:"brand"`
	SKU         string         `gorm:"index;size:255" json:"sku"`
	InStock     bool           `gorm:"default:true" json:"in_stock"`
	Rating      float32        `gorm:"type:decimal(3,2)" json:"rating"`
	ReviewCount int            `gorm:"default:0" json:"review_count"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Product model
func (Product) TableName() string {
	return "products"
}
