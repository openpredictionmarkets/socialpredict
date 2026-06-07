package models

import (
	"time"

	"gorm.io/gorm"
)

// MarketAccountingSnapshot stores durable display/read-model accounting values
// for a market. Raw markets and bets remain the canonical transaction source.
type MarketAccountingSnapshot struct {
	gorm.Model
	ID                 uint      `json:"id" gorm:"primaryKey"`
	MarketID           int64     `json:"marketId" gorm:"not null;uniqueIndex"`
	LastProbability    float64   `json:"lastProbability" gorm:"not null;default:0"`
	NetBetVolume       int64     `json:"netBetVolume" gorm:"not null;default:0"`
	MarketDust         int64     `json:"marketDust" gorm:"not null;default:0"`
	VolumeWithDust     int64     `json:"volumeWithDust" gorm:"not null;default:0"`
	UserCount          int       `json:"userCount" gorm:"not null;default:0"`
	BetCount           int       `json:"betCount" gorm:"not null;default:0"`
	LastProcessedBetID uint      `json:"lastProcessedBetId" gorm:"not null;default:0;index"`
	LastProcessedBetAt time.Time `json:"lastProcessedBetAt" gorm:"index"`
	GeneratedAt        time.Time `json:"generatedAt" gorm:"not null;index"`
	Source             string    `json:"source" gorm:"not null;default:read_model;size:32"`
}
