package models

import (
	"time"

	"gorm.io/gorm"
)

// ReportingVisibilitySettings controls which aggregate reporting endpoints are
// visible to logged-out visitors. User financial read models remain
// authenticated-only game transparency data.
type ReportingVisibilitySettings struct {
	gorm.Model
	ID                      uint   `gorm:"primaryKey"`
	Slug                    string `gorm:"not null;uniqueIndex;size:64"`
	SystemMetricsPublic     bool   `gorm:"not null;default:true"`
	GlobalLeaderboardPublic bool   `gorm:"not null;default:true"`
	Version                 uint   `gorm:"not null;default:1"`
	UpdatedBy               string `gorm:"size:64"`
	CreatedAt               time.Time
	UpdatedAt               time.Time
}
