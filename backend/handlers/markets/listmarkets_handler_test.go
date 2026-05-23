package marketshandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"socialpredict/handlers"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
)

type listMarketsServiceMock struct {
	listMarketsResult []*dmarkets.Market
	listMarketsErr    error
	listByStatusErr   error
	detailsErrByID    map[int64]error
	overviews         map[int64]*dmarkets.MarketOverview

	capturedFilters dmarkets.ListFilters
	capturedStatus  string
	capturedPage    dmarkets.Page
}

func (m *listMarketsServiceMock) CreateMarket(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error) {
	return nil, nil
}

func (m *listMarketsServiceMock) SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error {
	return nil
}

func (m *listMarketsServiceMock) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	return nil, nil
}

func (m *listMarketsServiceMock) ListMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
	m.capturedFilters = filters
	return m.listMarketsResult, m.listMarketsErr
}

func (m *listMarketsServiceMock) SearchMarkets(ctx context.Context, query string, filters dmarkets.SearchFilters) (*dmarkets.SearchResults, error) {
	return nil, nil
}

func (m *listMarketsServiceMock) ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error {
	return nil
}

func (m *listMarketsServiceMock) ListByStatus(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error) {
	m.capturedStatus = status
	m.capturedPage = p
	return m.listMarketsResult, m.listByStatusErr
}

func (m *listMarketsServiceMock) GetMarketLeaderboard(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
	return nil, nil
}

func (m *listMarketsServiceMock) ProjectProbability(ctx context.Context, req dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error) {
	return nil, nil
}

func (m *listMarketsServiceMock) GetMarketDetails(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error) {
	if err, ok := m.detailsErrByID[marketID]; ok {
		return nil, err
	}
	if overview, ok := m.overviews[marketID]; ok {
		return overview, nil
	}
	return &dmarkets.MarketOverview{
		Market:  &dmarkets.Market{ID: marketID},
		Creator: &dmarkets.CreatorSummary{Username: "tester"},
	}, nil
}

func (m *listMarketsServiceMock) GetMarketBets(ctx context.Context, marketID int64) ([]*dmarkets.BetDisplayInfo, error) {
	return nil, nil
}

func (m *listMarketsServiceMock) GetPublicMarket(ctx context.Context, marketID int64) (*dmarkets.PublicMarket, error) {
	return nil, nil
}

func (m *listMarketsServiceMock) GetMarketPositions(ctx context.Context, marketID int64) (dmarkets.MarketPositions, error) {
	return nil, nil
}

func (m *listMarketsServiceMock) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	return nil, nil
}

func (m *listMarketsServiceMock) CalculateMarketVolume(ctx context.Context, marketID int64) (int64, error) {
	return 0, nil
}

func TestListMarketsHandlerFactoryUsesStatusListing(t *testing.T) {
	mockSvc := &listMarketsServiceMock{
		listMarketsResult: []*dmarkets.Market{
			{ID: 42, QuestionTitle: "Active Market", Status: "active"},
		},
		overviews: map[int64]*dmarkets.MarketOverview{
			42: {
				Market: &dmarkets.Market{ID: 42, QuestionTitle: "Active Market", Status: "active"},
				Creator: &dmarkets.CreatorSummary{
					Username: "tester",
				},
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v0/markets?status=active&limit=5&offset=3", nil)
	res := httptest.NewRecorder()

	ListMarketsHandlerFactory(mockSvc)(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	if mockSvc.capturedStatus != "active" {
		t.Fatalf("expected ListByStatus status %q, got %q", "active", mockSvc.capturedStatus)
	}

	if mockSvc.capturedPage != (dmarkets.Page{Limit: 5, Offset: 3}) {
		t.Fatalf("unexpected page: %+v", mockSvc.capturedPage)
	}

	var resp dto.ListMarketsResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Total != 1 || len(resp.Markets) != 1 || resp.Markets[0].Market.ID != 42 {
		t.Fatalf("unexpected response payload: %+v", resp)
	}
}

func TestListMarketsHandlerFactoryUsesDefaultListingFilters(t *testing.T) {
	mockSvc := &listMarketsServiceMock{
		listMarketsResult: []*dmarkets.Market{
			{ID: 7, QuestionTitle: "Any Market"},
		},
		overviews: map[int64]*dmarkets.MarketOverview{
			7: {
				Market: &dmarkets.Market{ID: 7, QuestionTitle: "Any Market"},
				Creator: &dmarkets.CreatorSummary{
					Username: "tester",
				},
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v0/markets?limit=abc&offset=-2", nil)
	res := httptest.NewRecorder()

	ListMarketsHandlerFactory(mockSvc)(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	expectedFilters := dmarkets.ListFilters{Limit: 50, Offset: 0}
	if mockSvc.capturedFilters != expectedFilters {
		t.Fatalf("expected filters %+v, got %+v", expectedFilters, mockSvc.capturedFilters)
	}
}

func TestListMarketsHandlerFactoryUsesFailureEnvelope(t *testing.T) {
	t.Run("invalid status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v0/markets?status=maybe", nil)
		res := httptest.NewRecorder()

		ListMarketsHandlerFactory(&listMarketsServiceMock{})(res, req)

		assertFailureEnvelope(t, res, http.StatusBadRequest, handlers.ReasonInvalidRequest)
	})

	t.Run("domain validation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v0/markets", nil)
		res := httptest.NewRecorder()

		ListMarketsHandlerFactory(&listMarketsServiceMock{listMarketsErr: dmarkets.ErrInvalidInput})(res, req)

		assertFailureEnvelope(t, res, http.StatusBadRequest, handlers.ReasonValidationFailed)
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v0/markets", nil)
		res := httptest.NewRecorder()

		ListMarketsHandlerFactory(&listMarketsServiceMock{})(res, req)

		assertFailureEnvelope(t, res, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
	})
}

func TestGetMarketsHandlerUsesFailureEnvelope(t *testing.T) {
	t.Run("invalid status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v0/markets/all?status=maybe", nil)
		res := httptest.NewRecorder()

		GetMarketsHandler(&listMarketsServiceMock{})(res, req)

		assertFailureEnvelope(t, res, http.StatusBadRequest, handlers.ReasonInvalidRequest)
	})

	t.Run("domain validation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v0/markets/all", nil)
		res := httptest.NewRecorder()

		GetMarketsHandler(&listMarketsServiceMock{listMarketsErr: dmarkets.ErrInvalidInput})(res, req)

		assertFailureEnvelope(t, res, http.StatusBadRequest, handlers.ReasonValidationFailed)
	})
}

func TestBuildMarketOverviewResponsesIncludesMarketIDInError(t *testing.T) {
	mockSvc := &listMarketsServiceMock{
		detailsErrByID: map[int64]error{
			7: errors.New("boom"),
		},
	}

	_, err := buildMarketOverviewResponses(context.Background(), mockSvc, []*dmarkets.Market{
		{ID: 7, QuestionTitle: "Any Market"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "market_id=7") {
		t.Fatalf("expected market id in error, got %q", err.Error())
	}
}
