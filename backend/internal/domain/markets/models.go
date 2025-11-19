package markets

import (
	"time"
)

// Market represents the core market domain model
type Market struct {
	ID                      int64
	QuestionTitle           string
	Description             string
	OutcomeType             string
	ResolutionDateTime      time.Time
	FinalResolutionDateTime time.Time
	ResolutionResult        string
	CreatorUsername         string
	YesLabel                string
	NoLabel                 string
	Status                  string
	CreatedAt               time.Time
	UpdatedAt               time.Time
	InitialProbability      float64
	UTCOffset               int
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
	Username         string
	MarketID         int64
	YesSharesOwned   int64
	NoSharesOwned    int64
	Value            int64
	TotalSpent       int64
	TotalSpentInPlay int64
	IsResolved       bool
	ResolutionResult string
}

// MarketPositions aggregates user positions for a market.
type MarketPositions []*UserPosition

// Bet represents a wager placed within a market.
type Bet struct {
	ID        uint
	Username  string
	MarketID  uint
	Amount    int64
	Outcome   string
	PlacedAt  time.Time
	CreatedAt time.Time
}

// PayoutPosition captures the resolved valuation per user for distribution.
type PayoutPosition struct {
	Username string
	Value    int64
}

// PublicMarket represents the public view of a market.
type PublicMarket struct {
	ID                      int64
	QuestionTitle           string
	Description             string
	OutcomeType             string
	ResolutionDateTime      time.Time
	FinalResolutionDateTime time.Time
	UTCOffset               int
	IsResolved              bool
	ResolutionResult        string
	InitialProbability      float64
	CreatorUsername         string
	CreatedAt               time.Time
	YesLabel                string
	NoLabel                 string
}
