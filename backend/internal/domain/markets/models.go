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

// MarketCreateRequest represents the data needed to create a new market
type MarketCreateRequest struct {
	QuestionTitle      string
	Description        string
	OutcomeType        string
	ResolutionDateTime time.Time
	YesLabel           string
	NoLabel            string
}
