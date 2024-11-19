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

func CreateBet(username string, marketID uint, amount int64, outcome string) Bet {
	return Bet{
		Username: username,
		MarketID: marketID,
		Amount:   amount,
		PlacedAt: time.Now(),
		Outcome:  outcome,
	}
}
