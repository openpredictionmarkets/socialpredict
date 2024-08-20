package models

import (
	"errors"
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
	Amount   int64     `json:"amount"`
	PlacedAt time.Time `json:"placedAt"`
	Outcome  string    `json:"outcome,omitempty"`
}
type Bets []Bet

// getMarketUsers returns the number of unique users for a given market
func GetNumMarketUsers(bets []Bet) int {
	userMap := make(map[string]bool)
	for _, bet := range bets {
		userMap[bet.Username] = true
	}

	return len(userMap)
}

func (bet *Bet) ValidateBuy(db *gorm.DB) error {
	var user User
	var market Market

	// Check if username exists
	if err := db.First(&user, "username = ?", bet.Username).Error; err != nil {
		return errors.New("invalid username")
	}

	// Check if market exists and is open
	if err := db.First(&market, "id = ? AND is_resolved = false", bet.MarketID).Error; err != nil {
		return errors.New("invalid or closed market")
	}

	// Check for valid amount: it should be greater than or equal to 1
	if bet.Amount < 1 {
		return errors.New("amount must be greater than or equal to 1 for buying")
	}

	// Validate bet outcome: it should be either 'YES' or 'NO'
	if bet.Outcome != "YES" && bet.Outcome != "NO" {
		return errors.New("bet outcome must be 'YES' or 'NO'")
	}

	return nil
}
