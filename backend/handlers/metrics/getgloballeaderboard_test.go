package metricshandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	analytics "socialpredict/internal/domain/analytics"
	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"
)

type leaderboardRepo struct {
	users    []models.User
	markets  []models.Market
	betsByID map[uint][]models.Bet
}

func (r *leaderboardRepo) ListUsers(ctx context.Context) ([]models.User, error) {
	return append([]models.User(nil), r.users...), nil
}

func (r *leaderboardRepo) ListMarkets(ctx context.Context) ([]models.Market, error) {
	return append([]models.Market(nil), r.markets...), nil
}

func (r *leaderboardRepo) ListBetsForMarket(ctx context.Context, marketID uint) ([]models.Bet, error) {
	return append([]models.Bet(nil), r.betsByID[marketID]...), nil
}

func (r *leaderboardRepo) ListBetsOrdered(context.Context) ([]models.Bet, error) {
	return []models.Bet{}, nil
}

func (r *leaderboardRepo) UserMarketPositions(context.Context, string) ([]positionsmath.MarketPosition, error) {
	return []positionsmath.MarketPosition{}, nil
}

func TestGetGlobalLeaderboardHandler_Success(t *testing.T) {
	now := time.Now()
	repo := &leaderboardRepo{
		users: []models.User{
			{PublicUser: models.PublicUser{Username: "alice"}},
			{PublicUser: models.PublicUser{Username: "bob"}},
		},
		markets: []models.Market{
			{
				ID:                1,
				CreatorUsername:   "alice",
				IsResolved:        true,
				ResolutionResult:  "YES",
				ResolutionDateTime: now.Add(24 * time.Hour),
			},
		},
		betsByID: map[uint][]models.Bet{
			1: {
				{Username: "alice", Outcome: "YES", Amount: 100, MarketID: 1, PlacedAt: now.Add(-2 * time.Hour)},
				{Username: "bob", Outcome: "NO", Amount: 50, MarketID: 1, PlacedAt: now.Add(-1 * time.Hour)},
			},
		},
	}

	svc := analytics.NewService(repo, nil)
	handler := GetGlobalLeaderboardHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/v0/global/leaderboard", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var payload []analytics.GlobalUserProfitability
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(payload) == 0 {
		t.Fatalf("expected non-empty leaderboard")
	}
}

type failingRepo struct{}

func (f failingRepo) ListUsers(context.Context) ([]models.User, error) {
	return nil, assertError("boom")
}

func (f failingRepo) ListMarkets(context.Context) ([]models.Market, error) {
	return nil, nil
}

func (f failingRepo) ListBetsForMarket(context.Context, uint) ([]models.Bet, error) {
	return nil, nil
}

func (f failingRepo) ListBetsOrdered(context.Context) ([]models.Bet, error) {
	return nil, nil
}

func (f failingRepo) UserMarketPositions(context.Context, string) ([]positionsmath.MarketPosition, error) {
	return nil, nil
}

type assertError string

func (e assertError) Error() string { return string(e) }

func TestGetGlobalLeaderboardHandler_Error(t *testing.T) {
	svc := analytics.NewService(failingRepo{}, nil)
	handler := GetGlobalLeaderboardHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/v0/global/leaderboard", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}
