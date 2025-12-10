package bets

import "time"

// PlaceRequest captures the inputs required to place a buy bet.
type PlaceRequest struct {
	Username string
	MarketID uint
	Amount   int64
	Outcome  string
}

// PlacedBet represents the bet that was successfully recorded.
type PlacedBet struct {
	Username string
	MarketID uint
	Amount   int64
	Outcome  string
	PlacedAt time.Time
}

// SellRequest represents a request to sell shares for credits.
type SellRequest struct {
	Username string
	MarketID uint
	Amount   int64 // credits requested
	Outcome  string
}

// SellResult summarises the sale that occurred.
type SellResult struct {
	Username      string
	MarketID      uint
	SharesSold    int64
	SaleValue     int64
	Dust          int64
	Outcome       string
	TransactionAt time.Time
}
