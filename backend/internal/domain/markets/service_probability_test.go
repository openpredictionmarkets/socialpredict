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
	createFunc                   func(context.Context, *markets.Market) error
	updateLabelsFunc             func(context.Context, int64, string, string) error
	listFunc                     func(context.Context, markets.ListFilters) ([]*markets.Market, error)
	listByStatusFunc             func(context.Context, string, markets.Page) ([]*markets.Market, error)
	searchFunc                   func(context.Context, string, markets.SearchFilters) ([]*markets.Market, error)
	deleteFunc                   func(context.Context, int64) error
	resolveMarketFunc            func(context.Context, int64, string) error
	getUserPositionFunc          func(context.Context, int64, string) (*markets.UserPosition, error)
	listMarketPositionsFunc      func(context.Context, int64) (markets.MarketPositions, error)
	listBetsForMarketFunc        func(context.Context, int64) ([]*markets.Bet, error)
	getByIDFunc                  func(context.Context, int64) (*markets.Market, error)
	calculatePayoutPositionsFunc func(context.Context, int64) ([]*markets.PayoutPosition, error)
	getPublicMarketFunc          func(context.Context, int64) (*markets.PublicMarket, error)
}

func newProjectionRepo(opts ...func(*projectionRepo)) *projectionRepo {
	repo := &projectionRepo{
		createFunc:       func(context.Context, *markets.Market) error { return errUnexpectedMarketsTestCall },
		updateLabelsFunc: func(context.Context, int64, string, string) error { return errUnexpectedMarketsTestCall },
		listFunc: func(context.Context, markets.ListFilters) ([]*markets.Market, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		listByStatusFunc: func(context.Context, string, markets.Page) ([]*markets.Market, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		searchFunc: func(context.Context, string, markets.SearchFilters) ([]*markets.Market, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		deleteFunc:        func(context.Context, int64) error { return errUnexpectedMarketsTestCall },
		resolveMarketFunc: func(context.Context, int64, string) error { return errUnexpectedMarketsTestCall },
		getUserPositionFunc: func(context.Context, int64, string) (*markets.UserPosition, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		listMarketPositionsFunc: func(context.Context, int64) (markets.MarketPositions, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		listBetsForMarketFunc: func(context.Context, int64) ([]*markets.Bet, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		getByIDFunc: func(context.Context, int64) (*markets.Market, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		calculatePayoutPositionsFunc: func(context.Context, int64) ([]*markets.PayoutPosition, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		getPublicMarketFunc: func(context.Context, int64) (*markets.PublicMarket, error) {
			return nil, errUnexpectedMarketsTestCall
		},
	}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

func withProjectionRepoMarket(market *markets.Market) func(*projectionRepo) {
	return func(repo *projectionRepo) {
		repo.getByIDFunc = func(context.Context, int64) (*markets.Market, error) {
			if market == nil {
				return nil, markets.ErrMarketNotFound
			}
			return market, nil
		}
	}
}

func withProjectionRepoBets(bets []*markets.Bet) func(*projectionRepo) {
	return func(repo *projectionRepo) {
		repo.listBetsForMarketFunc = func(context.Context, int64) ([]*markets.Bet, error) {
			return bets, nil
		}
	}
}

func (r *projectionRepo) Create(ctx context.Context, market *markets.Market) error {
	return r.createFunc(ctx, market)
}
func (r *projectionRepo) UpdateLabels(ctx context.Context, id int64, yesLabel string, noLabel string) error {
	return r.updateLabelsFunc(ctx, id, yesLabel, noLabel)
}
func (r *projectionRepo) List(ctx context.Context, filters markets.ListFilters) ([]*markets.Market, error) {
	return r.listFunc(ctx, filters)
}
func (r *projectionRepo) ListByStatus(ctx context.Context, status string, page markets.Page) ([]*markets.Market, error) {
	return r.listByStatusFunc(ctx, status, page)
}
func (r *projectionRepo) Search(ctx context.Context, query string, filters markets.SearchFilters) ([]*markets.Market, error) {
	return r.searchFunc(ctx, query, filters)
}
func (r *projectionRepo) Delete(ctx context.Context, id int64) error { return r.deleteFunc(ctx, id) }
func (r *projectionRepo) ResolveMarket(ctx context.Context, id int64, outcome string) error {
	return r.resolveMarketFunc(ctx, id, outcome)
}
func (r *projectionRepo) GetUserPosition(ctx context.Context, marketID int64, username string) (*markets.UserPosition, error) {
	return r.getUserPositionFunc(ctx, marketID, username)
}
func (r *projectionRepo) ListMarketPositions(ctx context.Context, marketID int64) (markets.MarketPositions, error) {
	return r.listMarketPositionsFunc(ctx, marketID)
}
func (r *projectionRepo) ListBetsForMarket(ctx context.Context, marketID int64) ([]*markets.Bet, error) {
	return r.listBetsForMarketFunc(ctx, marketID)
}
func (r *projectionRepo) GetByID(ctx context.Context, id int64) (*markets.Market, error) {
	return r.getByIDFunc(ctx, id)
}
func (r *projectionRepo) CalculatePayoutPositions(ctx context.Context, marketID int64) ([]*markets.PayoutPosition, error) {
	return r.calculatePayoutPositionsFunc(ctx, marketID)
}

func (r *projectionRepo) GetPublicMarket(ctx context.Context, marketID int64) (*markets.PublicMarket, error) {
	return r.getPublicMarketFunc(ctx, marketID)
}

type projectionClock struct{ nowFunc func() time.Time }

func newProjectionClock(now time.Time) projectionClock {
	return projectionClock{nowFunc: func() time.Time { return now }}
}

func (c projectionClock) Now() time.Time {
	if c.nowFunc == nil {
		return time.Time{}
	}
	return c.nowFunc()
}

func TestProjectProbability_ComputesProjection(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	bets := []*markets.Bet{
		{Username: "alice", MarketID: 55, Amount: 100, Outcome: "YES", PlacedAt: createdAt.Add(5 * time.Minute), CreatedAt: createdAt.Add(5 * time.Minute)},
		{Username: "bob", MarketID: 55, Amount: 100, Outcome: "NO", PlacedAt: createdAt.Add(10 * time.Minute), CreatedAt: createdAt.Add(10 * time.Minute)},
	}
	repo := newProjectionRepo(
		withProjectionRepoMarket(&markets.Market{
			ID:                 55,
			Status:             "active",
			CreatedAt:          createdAt,
			ResolutionDateTime: createdAt.Add(48 * time.Hour),
		}),
		withProjectionRepoBets(bets),
	)

	svc := markets.NewService(repo, nil, newProjectionClock(createdAt.Add(20*time.Minute)), markets.Config{})

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

	expected := wpam.ProjectNewProbabilityWPAM(createdAt, marketsToModel(bets), modelsBet(createdAt.Add(20*time.Minute), 55, 50, "YES"))
	if absDiff(projection.ProjectedProbability, expected.Probability) > 1e-6 {
		t.Fatalf("expected projected %v got %v", expected.Probability, projection.ProjectedProbability)
	}
}

func TestProjectProbability_InvalidOutcome(t *testing.T) {
	repo := newProjectionRepo(withProjectionRepoMarket(&markets.Market{ID: 1, Status: "active", CreatedAt: time.Now(), ResolutionDateTime: time.Now().Add(time.Hour)}))
	svc := markets.NewService(repo, nil, newProjectionClock(time.Now()), markets.Config{})

	_, err := svc.ProjectProbability(context.Background(), markets.ProbabilityProjectionRequest{MarketID: 1, Amount: 10, Outcome: "MAYBE"})
	requireInvalidInput(t, err)
}

// helpers for tests

func marketsToModel(bets []*markets.Bet) []models.Bet {
	return markets.ToModelBets(bets)
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
