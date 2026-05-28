package models

import (
	"time"

	"gorm.io/gorm"
)

type SocialShareSettings struct {
	gorm.Model
	ID                 uint   `gorm:"primaryKey"`
	Slug               string `gorm:"uniqueIndex;size:64"`
	SiteName           string `gorm:"size:80"`
	DefaultDescription string `gorm:"size:220"`
	DefaultImageURL    string `gorm:"size:500"`
	ImageAlt           string `gorm:"size:160"`
	Version            uint   `gorm:"default:1"`
	UpdatedBy          string `gorm:"size:64"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type SocialShareImage struct {
	gorm.Model
	ID          uint   `gorm:"primaryKey"`
	Slug        string `gorm:"uniqueIndex;size:64"`
	FileName    string `gorm:"size:255"`
	ContentType string `gorm:"size:64"`
	SizeBytes   int64
	Data        []byte `gorm:"type:bytea"`
	UpdatedBy   string `gorm:"size:64"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
