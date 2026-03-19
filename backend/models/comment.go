package models

import "gorm.io/gorm"

// Comment represents a user comment on a prediction market.
type Comment struct {
	gorm.Model
	ID       uint   `json:"id" gorm:"primary_key"`
	MarketID uint   `json:"marketId" gorm:"not null;index"`
	Username string `json:"username" gorm:"not null;index"`
	Content  string `json:"content" gorm:"type:text;not null"`
}

type Comments []Comment
