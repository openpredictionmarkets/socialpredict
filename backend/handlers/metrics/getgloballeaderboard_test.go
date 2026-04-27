package metricshandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"socialpredict/handlers"
	analytics "socialpredict/internal/domain/analytics"
	"socialpredict/internal/domain/boundary"
	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models/modelstesting"
)

type leaderboardRepo struct {
	users    []analytics.UserAccount
	markets  []analytics.MarketRecord
	betsByID map[uint][]boundary.Bet
}

func (r *leaderboardRepo) ListUsers(ctx context.Context) ([]analytics.UserAccount, error) {
	return append([]analytics.UserAccount(nil), r.users...), nil
}

func (r *leaderboardRepo) ListMarkets(ctx context.Context) ([]analytics.MarketRecord, error) {
	return append([]analytics.MarketRecord(nil), r.markets...), nil
}

func (r *leaderboardRepo) ListBetsForMarket(ctx context.Context, marketID uint) ([]boundary.Bet, error) {
	return append([]boundary.Bet(nil), r.betsByID[marketID]...), nil
}

func (r *leaderboardRepo) ListBetsOrdered(context.Context) ([]boundary.Bet, error) {
	return []boundary.Bet{}, nil
}

func (r *leaderboardRepo) UserMarketPositions(context.Context, string) ([]positionsmath.MarketPosition, error) {
	return []positionsmath.MarketPosition{}, nil
}

func TestGetGlobalLeaderboardHandler_Success(t *testing.T) {
	_ = modelstesting.SeedWPAMFromConfig(modelstesting.GenerateEconomicConfig())

	now := time.Now()
	repo := &leaderboardRepo{
		users: []analytics.UserAccount{
			{Username: "alice"},
			{Username: "bob"},
		},
		markets: []analytics.MarketRecord{
			{
				ID:               1,
				CreatedAt:        now.Add(-24 * time.Hour),
				IsResolved:       true,
				ResolutionResult: "YES",
			},
		},
		betsByID: map[uint][]boundary.Bet{
			1: {
				{Username: "alice", Outcome: "YES", Amount: 100, MarketID: 1, PlacedAt: now.Add(-2 * time.Hour)},
				{Username: "bob", Outcome: "NO", Amount: 50, MarketID: 1, PlacedAt: now.Add(-1 * time.Hour)},
			},
		},
	}

	svc := analytics.NewService(repo, analytics.Config{})
	handler := GetGlobalLeaderboardHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/v0/global/leaderboard", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var payload handlers.SuccessEnvelope[[]analytics.GlobalUserProfitability]
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !payload.OK {
		t.Fatalf("expected ok=true, got false")
	}
	if len(payload.Result) == 0 {
		t.Fatalf("expected non-empty leaderboard")
	}
}

type failingRepo struct{}

func (f failingRepo) ListUsers(context.Context) ([]analytics.UserAccount, error) {
	return nil, assertError("boom")
}

func (f failingRepo) ListMarkets(context.Context) ([]analytics.MarketRecord, error) {
	return nil, nil
}

func (f failingRepo) ListBetsForMarket(context.Context, uint) ([]boundary.Bet, error) {
	return nil, nil
}

func (f failingRepo) ListBetsOrdered(context.Context) ([]boundary.Bet, error) {
	return nil, nil
}

func (f failingRepo) UserMarketPositions(context.Context, string) ([]positionsmath.MarketPosition, error) {
	return nil, nil
}

type assertError string

func (e assertError) Error() string { return string(e) }

func TestGetGlobalLeaderboardHandler_Error(t *testing.T) {
	svc := analytics.NewService(failingRepo{}, analytics.Config{})
	handler := GetGlobalLeaderboardHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/v0/global/leaderboard", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	var payload handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal failure: %v", err)
	}
	if payload.Reason != string(handlers.ReasonInternalError) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonInternalError, payload.Reason)
	}
}
