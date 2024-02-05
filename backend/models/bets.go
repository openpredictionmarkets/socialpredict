package models

import (
	"time"

	"gorm.io/gorm"
)

type Bet struct {
	gorm.Model
	Action   string    `json:"action"`
	ID       int64     `json:"id" gorm:"primary_key"`
	Username string    `json:"username"`
	User     User      `gorm:"foreignKey:Username;references:Username"`
	MarketID int64     `json:"marketId"`
	Market   Market    `gorm:"foreignKey:ID;references:MarketID"`
	Amount   int64     `json:"amount"`
	PlacedAt time.Time `json:"placedAt"`
	Outcome  string    `json:"outcome,omitempty"`
}
