package dto

import (
	"time"
)

// MarketResponse represents the HTTP response for a market
type MarketResponse struct {
	ID                 int64     `json:"id"`
	QuestionTitle      string    `json:"questionTitle"`
	Description        string    `json:"description"`
	OutcomeType        string    `json:"outcomeType"`
	ResolutionDateTime time.Time `json:"resolutionDateTime"`
	CreatorUsername    string    `json:"creatorUsername"`
	YesLabel           string    `json:"yesLabel"`
	NoLabel            string    `json:"noLabel"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// CreateMarketResponse represents the HTTP response after creating a market
type CreateMarketResponse struct {
	ID                 int64     `json:"id"`
	QuestionTitle      string    `json:"questionTitle"`
	Description        string    `json:"description"`
	OutcomeType        string    `json:"outcomeType"`
	ResolutionDateTime time.Time `json:"resolutionDateTime"`
	CreatorUsername    string    `json:"creatorUsername"`
	YesLabel           string    `json:"yesLabel"`
	NoLabel            string    `json:"noLabel"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"createdAt"`
}

// MarketOverviewResponse represents enriched market data for list display
type MarketOverviewResponse struct {
	Market          *MarketResponse `json:"market"`
	Creator         interface{}     `json:"creator"` // User info - will be properly typed later
	LastProbability float64         `json:"lastProbability"`
	NumUsers        int             `json:"numUsers"`
	TotalVolume     int64           `json:"totalVolume"`
}

// SimpleListMarketsResponse represents the HTTP response for simple market listing
type SimpleListMarketsResponse struct {
	Markets []*MarketResponse `json:"markets"`
	Total   int               `json:"total"`
}

// ListMarketsResponse represents the HTTP response for listing markets with enriched data
type ListMarketsResponse struct {
	Markets []*MarketOverviewResponse `json:"markets"`
}

// MarketOverview represents backward compatibility type for market overview data
type MarketOverview struct {
	Market          interface{} `json:"market"`
	Creator         interface{} `json:"creator"`
	LastProbability float64     `json:"lastProbability"`
	NumUsers        int         `json:"numUsers"`
	TotalVolume     int64       `json:"totalVolume"`
	MarketDust      int64       `json:"marketDust"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}
