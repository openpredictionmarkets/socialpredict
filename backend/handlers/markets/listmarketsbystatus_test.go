package marketshandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
)

type mockMarketsService struct {
	listByStatusFn func(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error)
}

func (m *mockMarketsService) CreateMarket(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error) {
	return nil, nil
}

func (m *mockMarketsService) SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error {
	return nil
}

func (m *mockMarketsService) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	return nil, nil
}

func (m *mockMarketsService) ListMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
	return nil, nil
}

func (m *mockMarketsService) SearchMarkets(ctx context.Context, query string, filters dmarkets.SearchFilters) (*dmarkets.SearchResults, error) {
	return &dmarkets.SearchResults{}, nil
}

func (m *mockMarketsService) ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error {
	return nil
}

func (m *mockMarketsService) ListByStatus(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error) {
	if m.listByStatusFn != nil {
		return m.listByStatusFn(ctx, status, p)
	}

	return []*dmarkets.Market{
		{
			ID:                 1,
			QuestionTitle:      status + " market",
			Description:        "Test " + status,
			OutcomeType:        "BINARY",
			ResolutionDateTime: time.Now().Add(24 * time.Hour),
			CreatorUsername:    "tester",
			YesLabel:           "YES",
			NoLabel:            "NO",
			Status:             status,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
	}, nil
}

func (m *mockMarketsService) GetMarketLeaderboard(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
	return nil, nil
}

func (m *mockMarketsService) ProjectProbability(ctx context.Context, req dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error) {
	return nil, nil
}

func (m *mockMarketsService) GetMarketDetails(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error) {
	return nil, nil
}

func (m *mockMarketsService) GetMarketBets(ctx context.Context, marketID int64) ([]*dmarkets.BetDisplayInfo, error) {
	return nil, nil
}

func (m *mockMarketsService) GetMarketPositions(ctx context.Context, marketID int64) (dmarkets.MarketPositions, error) {
	return nil, nil
}

func (m *mockMarketsService) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	return nil, nil
}

func TestListActiveMarketsHandler(t *testing.T) {
	mockSvc := &mockMarketsService{}
	handler := ListActiveMarketsHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/v0/markets/active", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp ListMarketsStatusResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Status != "active" {
		t.Fatalf("expected status active, got %s", resp.Status)
	}

	if resp.Count != 1 || len(resp.Markets) != 1 {
		t.Fatalf("expected single market in response, got count=%d len=%d", resp.Count, len(resp.Markets))
	}
}

func TestListClosedMarketsHandler(t *testing.T) {
	mockSvc := &mockMarketsService{}
	handler := ListClosedMarketsHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/v0/markets/closed", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp ListMarketsStatusResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Status != "closed" {
		t.Fatalf("expected status closed, got %s", resp.Status)
	}

	if resp.Count != 1 || len(resp.Markets) != 1 {
		t.Fatalf("expected single market in response, got count=%d len=%d", resp.Count, len(resp.Markets))
	}
}

func TestListResolvedMarketsHandler(t *testing.T) {
	mockSvc := &mockMarketsService{}
	handler := ListResolvedMarketsHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/v0/markets/resolved", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp ListMarketsStatusResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Status != "resolved" {
		t.Fatalf("expected status resolved, got %s", resp.Status)
	}

	if resp.Count != 1 || len(resp.Markets) != 1 {
		t.Fatalf("expected single market in response, got count=%d len=%d", resp.Count, len(resp.Markets))
	}
}

func TestListMarketsHandlerEmptyResponse(t *testing.T) {
	mockSvc := &mockMarketsService{
		listByStatusFn: func(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error) {
			return []*dmarkets.Market{}, nil
		},
	}

	handler := ListActiveMarketsHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/v0/markets/active", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp ListMarketsStatusResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Count != 0 || len(resp.Markets) != 0 {
		t.Fatalf("expected empty response, got count=%d len=%d", resp.Count, len(resp.Markets))
	}
}

func TestListMarketsHandlerMethodNotAllowed(t *testing.T) {
	mockSvc := &mockMarketsService{}
	handler := ListActiveMarketsHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/v0/markets/active", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
