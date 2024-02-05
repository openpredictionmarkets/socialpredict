package marketshandlers

import "time"

type PublicResponseMarket struct {
	ID                      int64     `json:"id"`
	QuestionTitle           string    `json:"questionTitle"`
	Description             string    `json:"description"`
	OutcomeType             string    `json:"outcomeType"`
	ResolutionDateTime      time.Time `json:"resolutionDateTime"`
	FinalResolutionDateTime time.Time `json:"finalResolutionDateTime"`
	UTCOffset               int       `json:"utcOffset"`
	IsResolved              bool      `json:"isResolved"`
	ResolutionResult        string    `json:"resolutionResult"`
	InitialProbability      float64   `json:"initialProbability"`
	CreatorUsername         string    `json:"creatorUsername"`
}
