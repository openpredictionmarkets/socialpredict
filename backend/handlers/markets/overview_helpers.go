package marketshandlers

import (
	"context"

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

	var creator *dto.CreatorResponse
	if overview.Creator != nil {
		creator = &dto.CreatorResponse{
			Username: overview.Creator.Username,
		}
	}

	return &dto.MarketOverviewResponse{
		Market:          marketToResponse(overview.Market),
		Creator:         creator,
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
		CreatedAt:          market.CreatedAt,
		UpdatedAt:          market.UpdatedAt,
	}
}
