package markets_test

import (
	"context"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	"socialpredict/internal/domain/math/probabilities/wpam"
	"socialpredict/models"
)

type betsRepo struct {
	market   *markets.Market
	bets     []*markets.Bet
	listErr  error
	marketID int64
}

func (r *betsRepo) Create(context.Context, *markets.Market) error { panic("unexpected call") }
func (r *betsRepo) UpdateLabels(context.Context, int64, string, string) error {
	panic("unexpected call")
}
func (r *betsRepo) List(context.Context, markets.ListFilters) ([]*markets.Market, error) {
	panic("unexpected call")
}
func (r *betsRepo) ListByStatus(context.Context, string, markets.Page) ([]*markets.Market, error) {
	panic("unexpected call")
}
func (r *betsRepo) Search(context.Context, string, markets.SearchFilters) ([]*markets.Market, error) {
	panic("unexpected call")
}
func (r *betsRepo) Delete(context.Context, int64) error { panic("unexpected call") }

func (r *betsRepo) GetByID(ctx context.Context, id int64) (*markets.Market, error) {
	if r.market == nil || r.market.ID != id {
		return nil, markets.ErrMarketNotFound
	}
	return r.market, nil
}

func (r *betsRepo) ResolveMarket(context.Context, int64, string) error { panic("unexpected call") }
func (r *betsRepo) GetUserPosition(context.Context, int64, string) (*markets.UserPosition, error) {
	panic("unexpected call")
}

func (r *betsRepo) ListBetsForMarket(ctx context.Context, marketID int64) ([]*markets.Bet, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.bets, nil
}

func (r *betsRepo) CalculatePayoutPositions(context.Context, int64) ([]*markets.PayoutPosition, error) {
	panic("unexpected call")
}

type nopUserService struct{}

func (nopUserService) ValidateUserExists(context.Context, string) error { return nil }
func (nopUserService) ValidateUserBalance(context.Context, string, float64, float64) error {
	return nil
}
func (nopUserService) DeductBalance(context.Context, string, float64) error          { return nil }
func (nopUserService) ApplyTransaction(context.Context, string, int64, string) error { return nil }

type betsClock struct{ now time.Time }

func (c betsClock) Now() time.Time { return c.now }

func TestGetMarketBets_ReturnsProbabilities(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	bets := []*markets.Bet{
		{Username: "alice", Outcome: "YES", Amount: 10, PlacedAt: createdAt.Add(1 * time.Minute)},
		{Username: "bob", Outcome: "NO", Amount: 15, PlacedAt: createdAt.Add(3 * time.Minute)},
		{Username: "carol", Outcome: "YES", Amount: 5, PlacedAt: createdAt.Add(5 * time.Minute)},
	}

	repo := &betsRepo{
		market: &markets.Market{
			ID:        42,
			CreatedAt: createdAt,
		},
		bets: bets,
	}

	service := markets.NewService(repo, nopUserService{}, betsClock{now: createdAt}, markets.Config{})

	results, err := service.GetMarketBets(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetMarketBets returned error: %v", err)
	}

	if len(results) != len(bets) {
		t.Fatalf("expected %d bets, got %d", len(bets), len(results))
	}

	modelBets := make([]models.Bet, len(bets))
	for i, bet := range bets {
		modelBets[i] = models.Bet{
			Username: bet.Username,
			MarketID: uint(bet.MarketID),
			Amount:   bet.Amount,
			Outcome:  bet.Outcome,
			PlacedAt: bet.PlacedAt,
		}
	}

	probabilityChanges := wpam.CalculateMarketProbabilitiesWPAM(createdAt, modelBets)

	matchProbability := func(bet models.Bet) float64 {
		prob := probabilityChanges[0].Probability
		for _, change := range probabilityChanges {
			if change.Timestamp.After(bet.PlacedAt) {
				break
			}
			prob = change.Probability
		}
		return prob
	}

	for i, bet := range modelBets {
		res := results[i]
		if res.Username != bet.Username || res.Amount != bet.Amount || !res.PlacedAt.Equal(bet.PlacedAt) {
			t.Fatalf("unexpected bet display info at index %d: %+v", i, res)
		}
		wantProb := matchProbability(bet)
		if res.Probability != wantProb {
			t.Fatalf("expected probability %.6f, got %.6f", wantProb, res.Probability)
		}
	}
}

func TestGetMarketBets_EmptyWhenNoBets(t *testing.T) {
	createdAt := time.Now()
	repo := &betsRepo{
		market: &markets.Market{
			ID:        7,
			CreatedAt: createdAt,
		},
		bets: nil,
	}

	service := markets.NewService(repo, nopUserService{}, betsClock{now: createdAt}, markets.Config{})

	results, err := service.GetMarketBets(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetMarketBets returned error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty result, got %d items", len(results))
	}
}

func TestGetMarketBets_ValidatesInputAndMarket(t *testing.T) {
	repo := &betsRepo{}
	service := markets.NewService(repo, nopUserService{}, betsClock{now: time.Now()}, markets.Config{})

	if _, err := service.GetMarketBets(context.Background(), 0); err != markets.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	if _, err := service.GetMarketBets(context.Background(), 99); err != markets.ErrMarketNotFound {
		t.Fatalf("expected ErrMarketNotFound, got %v", err)
	}
}
