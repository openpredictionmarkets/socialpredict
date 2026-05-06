package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	marketshandlers "socialpredict/handlers/markets"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/security"

	"github.com/gorilla/mux"
)

type marketsAliasServiceMock struct{}

func (m *marketsAliasServiceMock) CreateMarket(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error) {
	return nil, nil
}

func (m *marketsAliasServiceMock) SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error {
	return nil
}

func (m *marketsAliasServiceMock) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	return nil, nil
}

func (m *marketsAliasServiceMock) ListMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
	return nil, nil
}

func (m *marketsAliasServiceMock) GetMarketDetails(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error) {
	now := time.Now().UTC()
	return &dmarkets.MarketOverview{
		Market: &dmarkets.Market{
			ID:                 marketID,
			QuestionTitle:      "Active market",
			Description:        "desc",
			OutcomeType:        "BINARY",
			ResolutionDateTime: now.Add(24 * time.Hour),
			CreatorUsername:    "alice",
			YesLabel:           "YES",
			NoLabel:            "NO",
			Status:             "active",
			CreatedAt:          now,
			UpdatedAt:          now,
		},
		Creator: &dmarkets.CreatorSummary{Username: "alice"},
	}, nil
}

func (m *marketsAliasServiceMock) SearchMarkets(ctx context.Context, query string, filters dmarkets.SearchFilters) (*dmarkets.SearchResults, error) {
	return nil, nil
}

func (m *marketsAliasServiceMock) ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error {
	return nil
}

func (m *marketsAliasServiceMock) ListByStatus(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error) {
	if status != "active" {
		return nil, dmarkets.ErrInvalidInput
	}

	now := time.Now().UTC()
	return []*dmarkets.Market{{
		ID:                 7,
		QuestionTitle:      "Active market",
		Description:        "desc",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now.Add(24 * time.Hour),
		CreatorUsername:    "alice",
		YesLabel:           "YES",
		NoLabel:            "NO",
		Status:             "active",
		CreatedAt:          now,
		UpdatedAt:          now,
	}}, nil
}

func (m *marketsAliasServiceMock) GetMarketLeaderboard(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
	return nil, nil
}

func (m *marketsAliasServiceMock) ProjectProbability(ctx context.Context, req dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error) {
	return nil, nil
}

func TestLegacyMarketAliasRoutesPrecedeMarketIDRoute(t *testing.T) {
	router := mux.NewRouter()
	securityMiddleware := security.NewSecurityService().SecurityMiddleware()
	marketsHandler := marketshandlers.NewHandler(&marketsAliasServiceMock{}, nil, security.NewSecurityService())

	router.Handle("/v0/markets", securityMiddleware(http.HandlerFunc(marketsHandler.ListMarkets))).Methods("GET")
	router.Handle("/v0/markets", securityMiddleware(http.HandlerFunc(marketsHandler.CreateMarket))).Methods("POST")
	router.Handle("/v0/markets/search", securityMiddleware(http.HandlerFunc(marketsHandler.SearchMarkets))).Methods("GET")
	router.Handle("/v0/markets/status/{status}", securityMiddleware(http.HandlerFunc(marketsHandler.ListByStatus))).Methods("GET")
	router.Handle("/v0/markets/status", securityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rWithStatus := mux.SetURLVars(r, map[string]string{"status": "all"})
		marketsHandler.ListByStatus(w, rWithStatus)
	}))).Methods("GET")
	router.Handle("/v0/markets/active", securityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Set("status", "active")
		r.URL.RawQuery = q.Encode()
		marketsHandler.ListMarkets(w, r)
	}))).Methods("GET")
	router.Handle("/v0/markets/closed", securityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Set("status", "closed")
		r.URL.RawQuery = q.Encode()
		marketsHandler.ListMarkets(w, r)
	}))).Methods("GET")
	router.Handle("/v0/markets/resolved", securityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Set("status", "resolved")
		r.URL.RawQuery = q.Encode()
		marketsHandler.ListMarkets(w, r)
	}))).Methods("GET")
	router.Handle("/v0/markets/{id}", securityMiddleware(http.HandlerFunc(marketsHandler.GetDetails))).Methods("GET")

	req := httptest.NewRequest(http.MethodGet, "/v0/markets/active", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 for /v0/markets/active, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp dto.ListMarketsResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if resp.Total != 1 || len(resp.Markets) != 1 || resp.Markets[0].Market.ID != 7 {
		t.Fatalf("unexpected alias response: %+v", resp)
	}
}
