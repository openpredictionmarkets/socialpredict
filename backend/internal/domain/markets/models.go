package markets

import (
	"time"
)

// Market represents the core market domain model
type Market struct {
	ID                 int64
	QuestionTitle      string
	Description        string
	OutcomeType        string
	ResolutionDateTime time.Time
	CreatorUsername    string
	YesLabel           string
	NoLabel            string
	Status             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// MarketCreateRequest represents the request to create a new market
type MarketCreateRequest struct {
	QuestionTitle      string
	Description        string
	OutcomeType        string
	ResolutionDateTime time.Time
	YesLabel           string
	NoLabel            string
}

// UserPosition represents a user's holdings within a market.
type UserPosition struct {
	Username       string
	MarketID       int64
	YesSharesOwned int64
	NoSharesOwned  int64
	Value          int64
}

// MarketPositions aggregates user positions for a market.
type MarketPositions []*UserPosition
