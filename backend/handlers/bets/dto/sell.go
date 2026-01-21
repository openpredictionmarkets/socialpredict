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
	Outcome       string    `json:"outcome"`
	TransactionAt time.Time `json:"transactionAt"`
}
