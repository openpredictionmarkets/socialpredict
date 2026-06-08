package dto

import "time"

// SellBetRequest captures the payload for selling a position.
type SellBetRequest struct {
	MarketID uint   `json:"marketId"`
	Amount   int64  `json:"amount"`
	Outcome  string `json:"outcome"`
}

// SellBetResponse returns details about the sale performed.
type SellBetResponse struct {
	Username      string    `json:"username"`
	MarketID      uint      `json:"marketId"`
	SharesSold    int64     `json:"sharesSold"`
	SaleValue     int64     `json:"saleValue"`
	Dust          int64     `json:"dust"`
	NetProceeds   int64     `json:"netProceeds"`
	Outcome       string    `json:"outcome"`
	TransactionAt time.Time `json:"transactionAt"`
}

// SellQuoteResponse previews a sell request before settlement.
type SellQuoteResponse struct {
	Username          string    `json:"username"`
	MarketID          uint      `json:"marketId"`
	Outcome           string    `json:"outcome"`
	RequestedCredits  int64     `json:"requestedCredits"`
	SharesSold        int64     `json:"sharesSold"`
	SaleValue         int64     `json:"saleValue"`
	Dust              int64     `json:"dust"`
	NetProceeds       int64     `json:"netProceeds"`
	MaxDust           int64     `json:"maxDust"`
	ValuePerShare     int64     `json:"valuePerShare"`
	DustCapCoverage   float64   `json:"dustCapCoverage"`
	Allowed           bool      `json:"allowed"`
	SuggestedAmounts  []int64   `json:"suggestedAmounts"`
	Message           string    `json:"message"`
	QuotedAt          time.Time `json:"quotedAt"`
	DustCapExceeded   bool      `json:"dustCapExceeded"`
	DustCapExceededBy int64     `json:"dustCapExceededBy"`
}
