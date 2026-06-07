package models

import (
	"time"

	"gorm.io/gorm"
)

// AnalyticsReadModelSnapshot stores durable display-only analytics payloads.
// Payloads are JSON blobs owned by the analytics read-model boundary.
type AnalyticsReadModelSnapshot struct {
	gorm.Model
	ID            uint      `json:"id" gorm:"primaryKey"`
	SnapshotKey   string    `json:"snapshotKey" gorm:"not null;uniqueIndex;size:128"`
	Kind          string    `json:"kind" gorm:"not null;index;size:64"`
	PayloadJSON   string    `json:"payloadJson" gorm:"not null;type:text"`
	GeneratedAt   time.Time `json:"generatedAt" gorm:"not null;index"`
	Source        string    `json:"source" gorm:"not null;default:read_model;size:32"`
	IsStale       bool      `json:"isStale" gorm:"not null;default:false;index"`
	StaleReason   string    `json:"staleReason" gorm:"not null;default:'';size:128"`
	MarkedStaleAt time.Time `json:"markedStaleAt" gorm:"index"`
}
