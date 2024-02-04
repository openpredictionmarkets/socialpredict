package models

import (
	"time"

	"gorm.io/gorm"
)

type Bet struct {
	gorm.Model
	Action   string    `json:"action"`
	ID       uint      `json:"id" gorm:"primary_key"`
	Username string    `json:"username"`
	User     User      `gorm:"foreignKey:Username;references:Username"`
	MarketID uint      `json:"marketId"`
	Market   Market    `gorm:"foreignKey:ID;references:MarketID"`
	Amount   uint      `json:"amount"`
	PlacedAt time.Time `json:"placedAt"`
	Outcome  string    `json:"outcome,omitempty"`
}
