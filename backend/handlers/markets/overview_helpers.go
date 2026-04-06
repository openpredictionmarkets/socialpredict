package marketshandlers

import (
	"context"
	"strings"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
)

type marketOverviewProvider interface {
	GetMarketDetails(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error)
}

func buildMarketOverviewResponses(ctx context.Context, provider marketOverviewProvider, markets []*dmarkets.Market) ([]*dto.MarketOverviewResponse, error) {
	if len(markets) == 0 {
		return []*dto.MarketOverviewResponse{}, nil
	}

	overviews := make([]*dto.MarketOverviewResponse, 0, len(markets))
	for _, market := range markets {
		details, err := provider.GetMarketDetails(ctx, market.ID)
		if err != nil {
			return nil, err
		}
		overviews = append(overviews, marketOverviewToResponse(details))
	}
	return overviews, nil
}

func marketOverviewToResponse(overview *dmarkets.MarketOverview) *dto.MarketOverviewResponse {
	if overview == nil {
		return &dto.MarketOverviewResponse{}
	}

	return &dto.MarketOverviewResponse{
		Market:          marketToResponse(overview.Market),
		Creator:         creatorResponseFromSummary(overview.Creator),
		LastProbability: overview.LastProbability,
		NumUsers:        overview.NumUsers,
		TotalVolume:     overview.TotalVolume,
		MarketDust:      overview.MarketDust,
	}
}

func marketToResponse(market *dmarkets.Market) *dto.MarketResponse {
	if market == nil {
		return &dto.MarketResponse{}
	}

	return &dto.MarketResponse{
		ID:                 market.ID,
		QuestionTitle:      market.QuestionTitle,
		Description:        market.Description,
		OutcomeType:        market.OutcomeType,
		ResolutionDateTime: market.ResolutionDateTime,
		CreatorUsername:    market.CreatorUsername,
		YesLabel:           market.YesLabel,
		NoLabel:            market.NoLabel,
		Status:             market.Status,
		IsResolved:         strings.EqualFold(market.Status, "resolved"),
		ResolutionResult:   market.ResolutionResult,
		CreatedAt:          market.CreatedAt,
		UpdatedAt:          market.UpdatedAt,
	}
}

func creatorResponseFromSummary(summary *dmarkets.CreatorSummary) *dto.CreatorResponse {
	if summary == nil {
		return nil
	}
	return &dto.CreatorResponse{
		Username:      summary.Username,
		PersonalEmoji: summary.PersonalEmoji,
		DisplayName:   summary.DisplayName,
	}
}

func publicMarketResponseFromDomain(market *dmarkets.Market) dto.PublicMarketResponse {
	if market == nil {
		return dto.PublicMarketResponse{}
	}
	return dto.PublicMarketResponse{
		ID:                      market.ID,
		QuestionTitle:           market.QuestionTitle,
		Description:             market.Description,
		OutcomeType:             market.OutcomeType,
		ResolutionDateTime:      market.ResolutionDateTime,
		FinalResolutionDateTime: market.FinalResolutionDateTime,
		UTCOffset:               market.UTCOffset,
		IsResolved:              strings.EqualFold(market.Status, "resolved"),
		ResolutionResult:        market.ResolutionResult,
		InitialProbability:      market.InitialProbability,
		CreatorUsername:         market.CreatorUsername,
		CreatedAt:               market.CreatedAt,
		YesLabel:                market.YesLabel,
		NoLabel:                 market.NoLabel,
	}
}

func probabilityChangesToResponse(points []dmarkets.ProbabilityPoint) []dto.ProbabilityChangeResponse {
	if len(points) == 0 {
		return []dto.ProbabilityChangeResponse{}
	}
	result := make([]dto.ProbabilityChangeResponse, len(points))
	for i, point := range points {
		result[i] = dto.ProbabilityChangeResponse{
			Probability: point.Probability,
			Timestamp:   point.Timestamp,
		}
	}
	return result
}
