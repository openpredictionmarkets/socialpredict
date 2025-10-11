package models

import (
	"time"

	"gorm.io/gorm"
)

type HomepageContent struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey"`
	Slug      string `gorm:"uniqueIndex;size:64"` // always "home"
	Title     string `gorm:"size:255"`
	Format    string `gorm:"size:16"` // "markdown" or "html"
	Markdown  string `gorm:"type:text"`
	HTML      string `gorm:"type:text"`
	Version   uint   `gorm:"default:1"` // optimistic locking
	UpdatedBy string `gorm:"size:64"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
