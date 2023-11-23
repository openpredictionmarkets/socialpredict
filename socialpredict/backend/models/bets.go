package models

import (
	"time"

	"gorm.io/gorm"
)

type Bet struct {
	gorm.Model
	ID       uint      `json:"id" gorm:"primary_key"`
	UserID   uint      `json:"userId" gorm:"foreignKey:UserID"`
	User     User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	MarketID uint      `json:"marketId" gorm:"foreignKey:MarketID"`
	Market   Market    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Amount   float64   `json:"amount"`
	PlacedAt time.Time `json:"placedAt"`
	Outcome  string    `json:"outcome,omitempty"`
}
