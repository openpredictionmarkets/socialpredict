package dto

import "time"

// PlaceBetRequest represents the incoming payload for placing a bet.
type PlaceBetRequest struct {
	MarketID uint   `json:"marketId"`
	Amount   int64  `json:"amount"`
	Outcome  string `json:"outcome"`
}

// PlaceBetResponse represents the bet returned to the client after creation.
type PlaceBetResponse struct {
	Username string    `json:"username"`
	MarketID uint      `json:"marketId"`
	Amount   int64     `json:"amount"`
	Outcome  string    `json:"outcome"`
	PlacedAt time.Time `json:"placedAt"`
}
