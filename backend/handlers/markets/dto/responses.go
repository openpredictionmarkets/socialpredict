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

// ListMarketsResponse represents the HTTP response for listing markets
type ListMarketsResponse struct {
	Markets []*MarketResponse `json:"markets"`
	Total   int               `json:"total"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}
