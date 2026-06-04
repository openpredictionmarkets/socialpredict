package models

import (
	"time"

	"gorm.io/gorm"
)

type MarketDiscoveryPage struct {
	gorm.Model
	ID                         uint   `gorm:"primaryKey"`
	Slug                       string `gorm:"not null;uniqueIndex;size:96"`
	Title                      string `gorm:"not null;size:160"`
	Description                string `gorm:"size:500"`
	PageType                   string `gorm:"not null;index;size:32"`
	PrimaryTagSlug             string `gorm:"index;size:64"`
	SearchScope                string `gorm:"not null;default:all;size:32"`
	FeaturedTopicsEnabled      bool   `gorm:"not null;default:false"`
	FeaturedMarketsEnabled     bool   `gorm:"not null;default:false"`
	SectionsEnabled            bool   `gorm:"not null;default:false"`
	DefaultRecommendationLimit int    `gorm:"not null;default:20"`
	CuratedRecommendationLimit int    `gorm:"not null;default:5"`
	IsPublished                bool   `gorm:"not null;default:true;index"`
	SortOrder                  int    `gorm:"not null;default:0;index"`
	Version                    uint   `gorm:"not null;default:1"`
	UpdatedBy                  string `gorm:"size:64"`
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

type MarketDiscoverySection struct {
	gorm.Model
	ID            uint   `gorm:"primaryKey"`
	PageID        uint   `gorm:"not null;index"`
	Slug          string `gorm:"not null;size:96"`
	Title         string `gorm:"not null;size:160"`
	Description   string `gorm:"size:500"`
	TagFilterSlug string `gorm:"index;size:64"`
	SortOrder     int    `gorm:"not null;default:0;index"`
	IsActive      bool   `gorm:"not null;default:true;index"`
}

type MarketDiscoveryPin struct {
	gorm.Model
	ID             uint   `gorm:"primaryKey"`
	ScopeType      string `gorm:"not null;index;size:32"`
	ScopeID        uint   `gorm:"not null;index"`
	PinType        string `gorm:"not null;index;size:32"`
	MarketID       *int64 `gorm:"index"`
	TargetPageID   *uint  `gorm:"index"`
	TargetPageSlug string `gorm:"index;size:96"`
	Label          string `gorm:"size:160"`
	SortOrder      int    `gorm:"not null;default:0;index"`
	CreatedBy      string `gorm:"size:64"`
}
