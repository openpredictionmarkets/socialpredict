package markets_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"socialpredict/internal/domain/boundary"
	markets "socialpredict/internal/domain/markets"
	"socialpredict/internal/domain/math/probabilities/wpam"
	dusers "socialpredict/internal/domain/users"
)

type betsRepo struct {
	createFunc                   func(context.Context, *markets.Market) error
	updateLabelsFunc             func(context.Context, int64, string, string) error
	listFunc                     func(context.Context, markets.ListFilters) ([]*markets.Market, error)
	listByStatusFunc             func(context.Context, string, markets.Page) ([]*markets.Market, error)
	searchFunc                   func(context.Context, string, markets.SearchFilters) ([]*markets.Market, error)
	deleteFunc                   func(context.Context, int64) error
	getByIDFunc                  func(context.Context, int64) (*markets.Market, error)
	resolveMarketFunc            func(context.Context, int64, string) error
	getUserPositionFunc          func(context.Context, int64, string) (*markets.UserPosition, error)
	listMarketPositionsFunc      func(context.Context, int64) (markets.MarketPositions, error)
	listBetsForMarketFunc        func(context.Context, int64) ([]*markets.Bet, error)
	calculatePayoutPositionsFunc func(context.Context, int64) ([]*markets.PayoutPosition, error)
	getPublicMarketFunc          func(context.Context, int64) (*markets.PublicMarket, error)
}

func newBetsRepo(opts ...func(*betsRepo)) *betsRepo {
	repo := &betsRepo{
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
		deleteFunc: func(context.Context, int64) error { return errUnexpectedMarketsTestCall },
		getByIDFunc: func(context.Context, int64) (*markets.Market, error) {
			return nil, markets.ErrMarketNotFound
		},
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

func withBetsRepoMarket(market *markets.Market) func(*betsRepo) {
	return func(repo *betsRepo) {
		repo.getByIDFunc = func(context.Context, int64) (*markets.Market, error) {
			if market == nil {
				return nil, markets.ErrMarketNotFound
			}
			return market, nil
		}
	}
}

func withBetsRepoBets(bets []*markets.Bet) func(*betsRepo) {
	return func(repo *betsRepo) {
		repo.listBetsForMarketFunc = func(context.Context, int64) ([]*markets.Bet, error) {
			return bets, nil
		}
	}
}

func withBetsRepoListError(err error) func(*betsRepo) {
	return func(repo *betsRepo) {
		repo.listBetsForMarketFunc = func(context.Context, int64) ([]*markets.Bet, error) {
			return nil, err
		}
	}
}

func withBetsRepoPositions(positions markets.MarketPositions) func(*betsRepo) {
	return func(repo *betsRepo) {
		repo.listMarketPositionsFunc = func(context.Context, int64) (markets.MarketPositions, error) {
			return positions, nil
		}
	}
}

func (r *betsRepo) Create(ctx context.Context, market *markets.Market) error {
	if r.createFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.createFunc(ctx, market)
}
func (r *betsRepo) UpdateLabels(ctx context.Context, id int64, yesLabel string, noLabel string) error {
	if r.updateLabelsFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.updateLabelsFunc(ctx, id, yesLabel, noLabel)
}
func (r *betsRepo) List(ctx context.Context, filters markets.ListFilters) ([]*markets.Market, error) {
	if r.listFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listFunc(ctx, filters)
}
func (r *betsRepo) ListByStatus(ctx context.Context, status string, page markets.Page) ([]*markets.Market, error) {
	if r.listByStatusFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listByStatusFunc(ctx, status, page)
}
func (r *betsRepo) Search(ctx context.Context, query string, filters markets.SearchFilters) ([]*markets.Market, error) {
	if r.searchFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.searchFunc(ctx, query, filters)
}
func (r *betsRepo) Delete(ctx context.Context, id int64) error {
	if r.deleteFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.deleteFunc(ctx, id)
}

func (r *betsRepo) GetByID(ctx context.Context, id int64) (*markets.Market, error) {
	if r.getByIDFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.getByIDFunc(ctx, id)
}

func (r *betsRepo) ResolveMarket(ctx context.Context, id int64, outcome string) error {
	if r.resolveMarketFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.resolveMarketFunc(ctx, id, outcome)
}
func (r *betsRepo) GetUserPosition(ctx context.Context, marketID int64, username string) (*markets.UserPosition, error) {
	if r.getUserPositionFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.getUserPositionFunc(ctx, marketID, username)
}

func (r *betsRepo) ListMarketPositions(ctx context.Context, marketID int64) (markets.MarketPositions, error) {
	if r.listMarketPositionsFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listMarketPositionsFunc(ctx, marketID)
}

func (r *betsRepo) ListBetsForMarket(ctx context.Context, marketID int64) ([]*markets.Bet, error) {
	if r.listBetsForMarketFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listBetsForMarketFunc(ctx, marketID)
}

func (r *betsRepo) CalculatePayoutPositions(ctx context.Context, marketID int64) ([]*markets.PayoutPosition, error) {
	if r.calculatePayoutPositionsFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.calculatePayoutPositionsFunc(ctx, marketID)
}

func (r *betsRepo) GetPublicMarket(ctx context.Context, marketID int64) (*markets.PublicMarket, error) {
	if r.getPublicMarketFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.getPublicMarketFunc(ctx, marketID)
}

type nopUserService struct {
	validateUserExistsFunc  func(context.Context, string) error
	validateUserBalanceFunc func(context.Context, string, int64, int64) error
	deductBalanceFunc       func(context.Context, string, int64) error
	applyTransactionFunc    func(context.Context, string, int64, string) error
	getPublicUserFunc       func(context.Context, string) (*dusers.PublicUser, error)
}

func newNopUserService(opts ...func(*nopUserService)) nopUserService {
	service := nopUserService{
		validateUserExistsFunc:  func(context.Context, string) error { return nil },
		validateUserBalanceFunc: func(context.Context, string, int64, int64) error { return nil },
		deductBalanceFunc:       func(context.Context, string, int64) error { return nil },
		applyTransactionFunc:    func(context.Context, string, int64, string) error { return nil },
		getPublicUserFunc:       func(context.Context, string) (*dusers.PublicUser, error) { return nil, nil },
	}
	for _, opt := range opts {
		opt(&service)
	}
	return service
}

func (s nopUserService) ValidateUserExists(ctx context.Context, username string) error {
	if s.validateUserExistsFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return s.validateUserExistsFunc(ctx, username)
}
func (s nopUserService) ValidateUserBalance(ctx context.Context, username string, amount int64, maxDebt int64) error {
	if s.validateUserBalanceFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return s.validateUserBalanceFunc(ctx, username, amount, maxDebt)
}
func (s nopUserService) DeductBalance(ctx context.Context, username string, amount int64) error {
	if s.deductBalanceFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return s.deductBalanceFunc(ctx, username, amount)
}
func (s nopUserService) ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error {
	if s.applyTransactionFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return s.applyTransactionFunc(ctx, username, amount, transactionType)
}
func (s nopUserService) GetPublicUser(ctx context.Context, username string) (*dusers.PublicUser, error) {
	if s.getPublicUserFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return s.getPublicUserFunc(ctx, username)
}

type betsClock struct{ nowFunc func() time.Time }

func newBetsClock(now time.Time) betsClock {
	return betsClock{nowFunc: func() time.Time { return now }}
}

func (c betsClock) Now() time.Time {
	if c.nowFunc == nil {
		return marketsTestTime()
	}
	return c.nowFunc()
}

func requireInvalidInput(t *testing.T, err error) {
	t.Helper()
	if !markets.IsInvalidInput(err) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func requireMarketNotFound(t *testing.T, err error) {
	t.Helper()
	if !markets.IsMarketNotFound(err) {
		t.Fatalf("expected ErrMarketNotFound, got %v", err)
	}
}

func requireUnauthorized(t *testing.T, err error) {
	t.Helper()
	if !markets.IsUnauthorized(err) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func marketBetsToBoundary(bets []*markets.Bet) []boundary.Bet {
	return markets.ToBoundaryBets(bets)
}

func requireBetDisplay(t *testing.T, got *markets.BetDisplayInfo, want boundary.Bet, wantProbability float64) {
	t.Helper()
	if got.Username != want.Username || got.Amount != want.Amount || !got.PlacedAt.Equal(want.PlacedAt) {
		t.Fatalf("unexpected bet display info: %+v", got)
	}
	if got.Probability != wantProbability {
		t.Fatalf("expected probability %.6f, got %.6f", wantProbability, got.Probability)
	}
}

func TestGetMarketBets_ReturnsProbabilities(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	bets := []*markets.Bet{
		{Username: "alice", Outcome: "YES", Amount: 10, PlacedAt: createdAt.Add(1 * time.Minute)},
		{Username: "bob", Outcome: "NO", Amount: 15, PlacedAt: createdAt.Add(3 * time.Minute)},
		{Username: "carol", Outcome: "YES", Amount: 5, PlacedAt: createdAt.Add(5 * time.Minute)},
	}

	repo := newBetsRepo(
		withBetsRepoMarket(&markets.Market{
			ID:        42,
			CreatedAt: createdAt,
		}),
		withBetsRepoBets(bets),
	)

	service := markets.NewService(repo, newNopUserService(), newBetsClock(createdAt), markets.Config{})

	results, err := service.GetMarketBets(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetMarketBets returned error: %v", err)
	}

	if len(results) != len(bets) {
		t.Fatalf("expected %d bets, got %d", len(bets), len(results))
	}

	boundaryBets := marketBetsToBoundary(bets)
	probabilityChanges := wpam.CalculateMarketProbabilitiesWPAM(createdAt, boundaryBets)

	matchProbability := func(bet boundary.Bet) float64 {
		prob := probabilityChanges[0].Probability
		for _, change := range probabilityChanges {
			if change.Timestamp.After(bet.PlacedAt) {
				break
			}
			prob = change.Probability
		}
		return prob
	}

	for i, bet := range boundaryBets {
		wantProb := matchProbability(bet)
		requireBetDisplay(t, results[i], bet, wantProb)
	}

	if err := (&betsRepo{}).Create(context.Background(), nil); !errors.Is(err, errUnexpectedMarketsTestCall) {
		t.Fatalf("expected zero-value repo to fail predictably, got %v", err)
	}
}

func TestGetMarketBets_EmptyWhenNoBets(t *testing.T) {
	createdAt := marketsTestTime()
	repo := newBetsRepo(
		withBetsRepoMarket(&markets.Market{
			ID:        7,
			CreatedAt: createdAt,
		}),
		withBetsRepoBets(nil),
	)

	service := markets.NewService(repo, newNopUserService(), newBetsClock(createdAt), markets.Config{})

	results, err := service.GetMarketBets(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetMarketBets returned error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty result, got %d items", len(results))
	}

	if got := (betsClock{}).Now(); !got.Equal(marketsTestTime()) {
		t.Fatalf("expected zero-value clock fallback, got %v", got)
	}
}

func TestGetMarketBets_ValidatesInputAndMarket(t *testing.T) {
	repo := newBetsRepo()
	service := markets.NewService(repo, newNopUserService(), newBetsClock(marketsTestTime()), markets.Config{})

	_, err := service.GetMarketBets(context.Background(), 0)
	requireInvalidInput(t, err)

	_, err = service.GetMarketBets(context.Background(), 99)
	requireMarketNotFound(t, err)

	if _, err := (nopUserService{}).GetPublicUser(context.Background(), "alice"); !errors.Is(err, errUnexpectedMarketsTestCall) {
		t.Fatalf("expected zero-value user service to fail predictably, got %v", err)
	}
}

func TestGetMarketPositions_ReturnsRepositoryData(t *testing.T) {
	repo := newBetsRepo(
		withBetsRepoMarket(&markets.Market{ID: 101}),
		withBetsRepoPositions(markets.MarketPositions{
			{
				Username:         "alice",
				MarketID:         101,
				YesSharesOwned:   5,
				NoSharesOwned:    0,
				Value:            120,
				TotalSpent:       200,
				TotalSpentInPlay: 0,
				IsResolved:       true,
				ResolutionResult: "YES",
			},
			{
				Username:         "bob",
				MarketID:         101,
				YesSharesOwned:   0,
				NoSharesOwned:    3,
				Value:            0,
				TotalSpent:       75,
				TotalSpentInPlay: 0,
				IsResolved:       true,
				ResolutionResult: "YES",
			},
		}),
	)
	svc := markets.NewService(repo, newNopUserService(), newBetsClock(marketsTestTime()), markets.Config{})

	out, err := svc.GetMarketPositions(context.Background(), 101)
	if err != nil {
		t.Fatalf("GetMarketPositions returned error: %v", err)
	}
	out = out.Normalize()
	if len(out) != 2 {
		t.Fatalf("expected 2 positions, got %d", len(out))
	}
	if out[0].Username != "alice" || out[0].TotalSpent != 200 || !out[0].IsResolved {
		t.Fatalf("unexpected first position: %+v", out[0])
	}
	if out[1].Username != "bob" || out[1].NoSharesOwned != 3 {
		t.Fatalf("unexpected second position: %+v", out[1])
	}

	if _, err := (&betsRepo{}).ListMarketPositions(context.Background(), 1); !errors.Is(err, errUnexpectedMarketsTestCall) {
		t.Fatalf("expected zero-value positions repo to fail predictably, got %v", err)
	}
}

func TestGetMarketPositions_ValidatesInputAndMarket(t *testing.T) {
	repo := newBetsRepo()
	svc := markets.NewService(repo, newNopUserService(), newBetsClock(marketsTestTime()), markets.Config{})

	_, err := svc.GetMarketPositions(context.Background(), 0)
	requireInvalidInput(t, err)

	_, err = svc.GetMarketPositions(context.Background(), 99)
	requireMarketNotFound(t, err)

	if err := (nopUserService{}).ApplyTransaction(context.Background(), "alice", 10, "test"); !errors.Is(err, errUnexpectedMarketsTestCall) {
		t.Fatalf("expected zero-value user service to fail predictably, got %v", err)
	}
}
