package marketshandlers

import (
	"context"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
)

// MockService provides a reusable test double for markets service interactions.
type MockService struct {
	ListByStatusFn        func(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error)
	ListLifecycleFn       func(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error)
	MarketLeaderboardFn   func(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error)
	MarketSummaryFn       func(ctx context.Context, marketID int64) (*dmarkets.MarketSummaryReadModel, error)
	CreateMarketGroupFn   func(ctx context.Context, req dmarkets.MarketGroupCreateRequest, creatorUsername string) (*dmarkets.MarketGroup, error)
	MarketGroupOverviewFn func(ctx context.Context, groupID int64) (*dmarkets.MarketGroupOverview, error)
	ResolveMarketGroupFn  func(ctx context.Context, groupID int64, req dmarkets.MarketGroupResolveRequest, username string) (*dmarkets.MarketGroup, error)
	MarketGroupLookupFn   func(ctx context.Context, marketID int64) (*dmarkets.MarketGroup, error)
	DetailsCalls          int
}

func (m *MockService) CreateMarket(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error) {
	return nil, nil
}

func (m *MockService) CreateMarketGroup(ctx context.Context, req dmarkets.MarketGroupCreateRequest, creatorUsername string) (*dmarkets.MarketGroup, error) {
	if m.CreateMarketGroupFn != nil {
		return m.CreateMarketGroupFn(ctx, req, creatorUsername)
	}
	return nil, dmarkets.ErrInvalidInput
}

func (m *MockService) GetMarketGroupOverview(ctx context.Context, groupID int64) (*dmarkets.MarketGroupOverview, error) {
	if m.MarketGroupOverviewFn != nil {
		return m.MarketGroupOverviewFn(ctx, groupID)
	}
	return nil, dmarkets.ErrMarketGroupNotFound
}

func (m *MockService) ResolveMarketGroup(ctx context.Context, groupID int64, req dmarkets.MarketGroupResolveRequest, username string) (*dmarkets.MarketGroup, error) {
	if m.ResolveMarketGroupFn != nil {
		return m.ResolveMarketGroupFn(ctx, groupID, req, username)
	}
	return nil, dmarkets.ErrInvalidInput
}

func (m *MockService) GetMarketGroupForMarket(ctx context.Context, marketID int64) (*dmarkets.MarketGroup, error) {
	if m.MarketGroupLookupFn != nil {
		return m.MarketGroupLookupFn(ctx, marketID)
	}
	return nil, dmarkets.ErrMarketGroupNotFound
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

func (m *MockService) ListLifecycleMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
	if m.ListLifecycleFn != nil {
		return m.ListLifecycleFn(ctx, filters)
	}
	return []*dmarkets.Market{}, nil
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
	m.DetailsCalls += 1
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

func (m *MockService) GetMarketSummaryReadModel(ctx context.Context, marketID int64) (*dmarkets.MarketSummaryReadModel, error) {
	if m.MarketSummaryFn != nil {
		return m.MarketSummaryFn(ctx, marketID)
	}
	now := time.Now()
	return &dmarkets.MarketSummaryReadModel{
		Market: &dmarkets.Market{
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
		},
		Creator: &dmarkets.CreatorSummary{Username: "testuser"},
		Accounting: dmarkets.MarketAccountingSnapshot{
			MarketID:           marketID,
			GeneratedAt:        now,
			ProbabilityChanges: []dmarkets.ProbabilityPoint{},
			LastProbability:    0.5,
			UserCount:          3,
			VolumeWithDust:     1000,
			MarketDust:         50,
			Source:             "read_model",
		},
	}, nil
}

func (m *MockService) GetMarketBets(ctx context.Context, marketID int64) ([]*dmarkets.BetDisplayInfo, error) {
	return []*dmarkets.BetDisplayInfo{}, nil
}

func (m *MockService) GetMarketBetsPage(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.BetDisplayInfo, error) {
	return []*dmarkets.BetDisplayInfo{}, nil
}

func (m *MockService) GetMarketPositions(ctx context.Context, marketID int64) (dmarkets.MarketPositions, error) {
	return nil, nil
}

func (m *MockService) GetMarketPositionsPage(ctx context.Context, marketID int64, p dmarkets.Page) (dmarkets.MarketPositions, error) {
	return nil, nil
}

func (m *MockService) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	return nil, nil
}

func (m *MockService) CalculateMarketVolume(ctx context.Context, marketID int64) (int64, error) {
	return 0, nil
}

func (m *MockService) GetPublicMarket(ctx context.Context, marketID int64) (*dmarkets.PublicMarket, error) {
	return &dmarkets.PublicMarket{ID: marketID}, nil
}

func (m *MockService) GetShareMetadata(ctx context.Context, marketID int64, config dmarkets.ShareMetadataConfig) (*dmarkets.ShareMetadata, error) {
	return &dmarkets.ShareMetadata{
		MarketID:     marketID,
		Title:        "Test Market | SocialPredict",
		Description:  "Test market description",
		CanonicalURL: "https://example.test/markets/1",
		ImageURL:     "https://example.test/logo512.png",
		PublicStatus: dmarkets.MarketStatusActive,
		SiteName:     "SocialPredict",
		Shareable:    true,
	}, nil
}
