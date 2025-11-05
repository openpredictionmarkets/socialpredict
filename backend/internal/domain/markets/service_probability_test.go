package markets_test

import (
	"context"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	"socialpredict/internal/domain/math/probabilities/wpam"
	"socialpredict/models"
)

type projectionRepo struct {
	market *markets.Market
	bets   []*markets.Bet
}

func (r *projectionRepo) Create(context.Context, *markets.Market) error { panic("unexpected call") }
func (r *projectionRepo) UpdateLabels(context.Context, int64, string, string) error {
	panic("unexpected call")
}
func (r *projectionRepo) List(context.Context, markets.ListFilters) ([]*markets.Market, error) {
	panic("unexpected call")
}
func (r *projectionRepo) ListByStatus(context.Context, string, markets.Page) ([]*markets.Market, error) {
	panic("unexpected call")
}
func (r *projectionRepo) Search(context.Context, string, markets.SearchFilters) ([]*markets.Market, error) {
	panic("unexpected call")
}
func (r *projectionRepo) Delete(context.Context, int64) error { panic("unexpected call") }
func (r *projectionRepo) ResolveMarket(context.Context, int64, string) error {
	panic("unexpected call")
}
func (r *projectionRepo) GetUserPosition(context.Context, int64, string) (*markets.UserPosition, error) {
	panic("unexpected call")
}
func (r *projectionRepo) ListMarketPositions(context.Context, int64) (markets.MarketPositions, error) {
	panic("unexpected call")
}
func (r *projectionRepo) ListBetsForMarket(ctx context.Context, marketID int64) ([]*markets.Bet, error) {
	return r.bets, nil
}
func (r *projectionRepo) GetByID(ctx context.Context, id int64) (*markets.Market, error) {
	if r.market == nil || r.market.ID != id {
		return nil, markets.ErrMarketNotFound
	}
	return r.market, nil
}
func (r *projectionRepo) CalculatePayoutPositions(context.Context, int64) ([]*markets.PayoutPosition, error) {
	panic("unexpected call")
}

func (r *projectionRepo) GetPublicMarket(context.Context, int64) (*markets.PublicMarket, error) {
	panic("unexpected call")
}

type projectionClock struct{ now time.Time }

func (c projectionClock) Now() time.Time { return c.now }

func TestProjectProbability_ComputesProjection(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	repo := &projectionRepo{
		market: &markets.Market{
			ID:                 55,
			Status:             "active",
			CreatedAt:          createdAt,
			ResolutionDateTime: createdAt.Add(48 * time.Hour),
		},
		bets: []*markets.Bet{
			{Username: "alice", MarketID: 55, Amount: 100, Outcome: "YES", PlacedAt: createdAt.Add(5 * time.Minute), CreatedAt: createdAt.Add(5 * time.Minute)},
			{Username: "bob", MarketID: 55, Amount: 100, Outcome: "NO", PlacedAt: createdAt.Add(10 * time.Minute), CreatedAt: createdAt.Add(10 * time.Minute)},
		},
	}

	svc := markets.NewService(repo, nil, projectionClock{now: createdAt.Add(20 * time.Minute)}, markets.Config{})

	projection, err := svc.ProjectProbability(context.Background(), markets.ProbabilityProjectionRequest{
		MarketID: 55,
		Amount:   50,
		Outcome:  "YES",
	})
	if err != nil {
		t.Fatalf("ProjectProbability returned error: %v", err)
	}

	if projection.CurrentProbability <= 0 || projection.CurrentProbability >= 1 {
		t.Fatalf("unexpected current probability: %v", projection.CurrentProbability)
	}

	expected := wpam.ProjectNewProbabilityWPAM(createdAt, marketsToModel(repo.bets), modelsBet(createdAt.Add(20*time.Minute), 55, 50, "YES"))
	if absDiff(projection.ProjectedProbability, expected.Probability) > 1e-6 {
		t.Fatalf("expected projected %v got %v", expected.Probability, projection.ProjectedProbability)
	}
}

func TestProjectProbability_InvalidOutcome(t *testing.T) {
	repo := &projectionRepo{market: &markets.Market{ID: 1, Status: "active", CreatedAt: time.Now(), ResolutionDateTime: time.Now().Add(time.Hour)}}
	svc := markets.NewService(repo, nil, projectionClock{now: time.Now()}, markets.Config{})

	if _, err := svc.ProjectProbability(context.Background(), markets.ProbabilityProjectionRequest{MarketID: 1, Amount: 10, Outcome: "MAYBE"}); err != markets.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

// helpers for tests

func marketsToModel(bets []*markets.Bet) []models.Bet {
	result := make([]models.Bet, len(bets))
	for i, b := range bets {
		result[i] = models.Bet{
			Username: b.Username,
			MarketID: uint(b.MarketID),
			Amount:   b.Amount,
			Outcome:  b.Outcome,
			PlacedAt: b.PlacedAt,
		}
	}
	return result
}

func modelsBet(placed time.Time, marketID int64, amount int64, outcome string) models.Bet {
	return models.Bet{Username: "preview", MarketID: uint(marketID), Amount: amount, Outcome: outcome, PlacedAt: placed}
}

func absDiff(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}
