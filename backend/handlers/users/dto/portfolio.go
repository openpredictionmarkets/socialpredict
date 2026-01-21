package dto

import "time"

// PortfolioItemResponse represents a single market entry in the user's portfolio response.
type PortfolioItemResponse struct {
	MarketID       uint      `json:"marketId"`
	QuestionTitle  string    `json:"questionTitle"`
	YesSharesOwned int64     `json:"yesSharesOwned"`
	NoSharesOwned  int64     `json:"noSharesOwned"`
	LastBetPlaced  time.Time `json:"lastBetPlaced"`
}

// PortfolioResponse represents the payload returned for a user's portfolio request.
type PortfolioResponse struct {
	PortfolioItems   []PortfolioItemResponse `json:"portfolioItems"`
	TotalSharesOwned int64                   `json:"totalSharesOwned"`
}

