package models

import "time"

func CreateBet(username string, marketID uint, amount int64, outcome string) Bet {
	return Bet{
		Username: username,
		MarketID: marketID,
		Amount:   amount,
		PlacedAt: time.Now(),
		Outcome:  outcome,
	}
}
