package marketpublicresponse

import (
	"context"
	"errors"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
)

// PublicResponseMarket mirrors the fields exposed by the legacy public market response.
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
	CreatedAt               time.Time `json:"createdAt"`
	YesLabel                string    `json:"yesLabel"`
	NoLabel                 string    `json:"noLabel"`
}

// GetPublicResponseMarket fetches a market's public data through the markets service.
func GetPublicResponseMarket(ctx context.Context, svc dmarkets.ServiceInterface, marketID int64) (*PublicResponseMarket, error) {
	if svc == nil {
		return nil, errors.New("market service is nil")
	}

	market, err := svc.GetPublicMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}

	return &PublicResponseMarket{
		ID:                      market.ID,
		QuestionTitle:           market.QuestionTitle,
		Description:             market.Description,
		OutcomeType:             market.OutcomeType,
		ResolutionDateTime:      market.ResolutionDateTime,
		FinalResolutionDateTime: market.FinalResolutionDateTime,
		UTCOffset:               market.UTCOffset,
		IsResolved:              market.IsResolved,
		ResolutionResult:        market.ResolutionResult,
		InitialProbability:      market.InitialProbability,
		CreatorUsername:         market.CreatorUsername,
		CreatedAt:               market.CreatedAt,
		YesLabel:                market.YesLabel,
		NoLabel:                 market.NoLabel,
	}, nil
}
