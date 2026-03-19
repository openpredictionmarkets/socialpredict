package models

import "gorm.io/gorm"

// Notification is an in-app alert for a user.
// Currently only "market_resolved" notifications are generated
// (when a market the user bet on gets resolved).
type Notification struct {
	gorm.Model
	ID       uint   `json:"id" gorm:"primary_key"`
	Username string `json:"username" gorm:"not null;index"`
	Type     string `json:"type" gorm:"not null"`  // e.g. "market_resolved"
	MarketID uint   `json:"marketId" gorm:"index"` // 0 means not market-specific
	Message  string `json:"message" gorm:"type:text;not null"`
	IsRead   bool   `json:"isRead" gorm:"default:false"`
}

type Notifications []Notification
