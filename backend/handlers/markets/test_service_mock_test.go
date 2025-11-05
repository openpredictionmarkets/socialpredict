package marketshandlers

import (
	"context"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
)

// MockService provides a reusable test double for markets service interactions.
type MockService struct {
	ListByStatusFn      func(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error)
	MarketLeaderboardFn func(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error)
}

func (m *MockService) CreateMarket(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error) {
	return nil, nil
}

func (m *MockService) SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error {
	return nil
}

func (m *MockService) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	return nil, nil
}

func (m *MockService) ListMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
	return nil, nil
}

func (m *MockService) SearchMarkets(ctx context.Context, query string, filters dmarkets.SearchFilters) (*dmarkets.SearchResults, error) {
	return &dmarkets.SearchResults{
		PrimaryResults:  []*dmarkets.Market{},
		FallbackResults: []*dmarkets.Market{},
		Query:           query,
		PrimaryStatus:   filters.Status,
	}, nil
}

func (m *MockService) ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error {
	return nil
}

func (m *MockService) ListByStatus(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error) {
	if m.ListByStatusFn != nil {
		return m.ListByStatusFn(ctx, status, p)
	}

	now := time.Now()
	return []*dmarkets.Market{
		{
			ID:                 1,
			QuestionTitle:      status + " Market",
			Description:        "Test " + status + " market",
			OutcomeType:        "BINARY",
			ResolutionDateTime: now.Add(24 * time.Hour),
			CreatorUsername:    "testuser",
			YesLabel:           "YES",
			NoLabel:            "NO",
			Status:             status,
			CreatedAt:          now,
			UpdatedAt:          now,
		},
	}, nil
}

func (m *MockService) GetMarketLeaderboard(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
	if m.MarketLeaderboardFn != nil {
		return m.MarketLeaderboardFn(ctx, marketID, p)
	}
	return []*dmarkets.LeaderboardRow{}, nil
}

func (m *MockService) ProjectProbability(ctx context.Context, req dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error) {
	return &dmarkets.ProbabilityProjection{
		CurrentProbability:   0.5,
		ProjectedProbability: 0.6,
	}, nil
}

func (m *MockService) GetMarketDetails(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error) {
	now := time.Now()

	market := &dmarkets.Market{
		ID:                 marketID,
		QuestionTitle:      "Test Market",
		Description:        "Test market description",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now.Add(24 * time.Hour),
		CreatorUsername:    "testuser",
		YesLabel:           "YES",
		NoLabel:            "NO",
		Status:             "active",
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	var marketDust int64
	var totalVolume int64
	var numUsers int

	if marketID == 1 {
		marketDust = 50
		totalVolume = 1000
		numUsers = 3
	}

	return &dmarkets.MarketOverview{
		Market:             market,
		Creator:            &dmarkets.CreatorSummary{Username: "testuser"},
		ProbabilityChanges: []dmarkets.ProbabilityPoint{},
		LastProbability:    0,
		NumUsers:           numUsers,
		TotalVolume:        totalVolume,
		MarketDust:         marketDust,
	}, nil
}

func (m *MockService) GetMarketBets(ctx context.Context, marketID int64) ([]*dmarkets.BetDisplayInfo, error) {
	return []*dmarkets.BetDisplayInfo{}, nil
}

func (m *MockService) GetMarketPositions(ctx context.Context, marketID int64) (dmarkets.MarketPositions, error) {
	return nil, nil
}

func (m *MockService) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	return nil, nil
}

func (m *MockService) CalculateMarketVolume(ctx context.Context, marketID int64) (int64, error) {
	return 0, nil
}
