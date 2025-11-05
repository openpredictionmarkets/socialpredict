package marketshandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
)

type searchServiceMock struct {
	result          *dmarkets.SearchResults
	err             error
	capturedQuery   string
	capturedFilters dmarkets.SearchFilters
}

func (m *searchServiceMock) CreateMarket(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error) {
	return nil, nil
}

func (m *searchServiceMock) SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error {
	return nil
}

func (m *searchServiceMock) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	return nil, nil
}

func (m *searchServiceMock) ListMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
	return nil, nil
}

func (m *searchServiceMock) SearchMarkets(ctx context.Context, query string, filters dmarkets.SearchFilters) (*dmarkets.SearchResults, error) {
	m.capturedQuery = query
	m.capturedFilters = filters
	return m.result, m.err
}

func (m *searchServiceMock) ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error {
	return nil
}

func (m *searchServiceMock) ListByStatus(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error) {
	return nil, nil
}

func (m *searchServiceMock) GetMarketLeaderboard(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
	return nil, nil
}

func (m *searchServiceMock) ProjectProbability(ctx context.Context, req dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error) {
	return nil, nil
}

func (m *searchServiceMock) GetMarketDetails(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error) {
	return nil, nil
}

func (m *searchServiceMock) GetMarketBets(ctx context.Context, marketID int64) ([]*dmarkets.BetDisplayInfo, error) {
	return nil, nil
}

func (m *searchServiceMock) GetPublicMarket(ctx context.Context, marketID int64) (*dmarkets.PublicMarket, error) {
	return &dmarkets.PublicMarket{ID: marketID}, nil
}

func (m *searchServiceMock) GetMarketPositions(ctx context.Context, marketID int64) (dmarkets.MarketPositions, error) {
	return nil, nil
}

func (m *searchServiceMock) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	return nil, nil
}

func (m *searchServiceMock) CalculateMarketVolume(ctx context.Context, marketID int64) (int64, error) {
	return 0, nil
}

func TestSearchMarketsHandlerSuccess(t *testing.T) {
	mockResult := &dmarkets.SearchResults{
		PrimaryResults: []*dmarkets.Market{
			{ID: 1, QuestionTitle: "Test Market", CreatorUsername: "tester"},
		},
		Query:         "bitcoin",
		PrimaryStatus: "active",
		PrimaryCount:  1,
		TotalCount:    1,
	}

	mockSvc := &searchServiceMock{result: mockResult}
	handler := SearchMarketsHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/v0/markets/search?q=bitcoin&status=active&limit=5&offset=2", nil)
	res := httptest.NewRecorder()

	handler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	if mockSvc.capturedQuery != "bitcoin" {
		t.Fatalf("expected query to be sanitized value 'bitcoin', got %s", mockSvc.capturedQuery)
	}

	if mockSvc.capturedFilters.Status != "active" || mockSvc.capturedFilters.Limit != 5 || mockSvc.capturedFilters.Offset != 2 {
		t.Fatalf("unexpected filters captured: %+v", mockSvc.capturedFilters)
	}

	var resp dto.SearchResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.TotalCount != 1 || resp.PrimaryCount != 1 {
		t.Fatalf("expected counts to be 1, got total=%d primary=%d", resp.TotalCount, resp.PrimaryCount)
	}
}

func TestSearchMarketsHandlerValidation(t *testing.T) {
	mockSvc := &searchServiceMock{}
	handler := SearchMarketsHandler(mockSvc)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v0/markets/search", nil)
		rr := httptest.NewRecorder()

		handler(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected %d, got %d", http.StatusMethodNotAllowed, rr.Code)
		}
	})

	t.Run("missing query parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v0/markets/search", nil)
		rr := httptest.NewRecorder()

		handler(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("domain service error surfaces as server error", func(t *testing.T) {
		mockSvc.err = errors.New("boom")
		req := httptest.NewRequest(http.MethodGet, "/v0/markets/search?q=test", nil)
		rr := httptest.NewRecorder()

		handler(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected %d, got %d", http.StatusInternalServerError, rr.Code)
		}
	})
}
