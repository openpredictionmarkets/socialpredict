package markets_test

import (
	"context"
	"errors"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	users "socialpredict/internal/domain/users"
)

type resolveRepo struct {
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

func newResolveRepo(opts ...func(*resolveRepo)) *resolveRepo {
	repo := &resolveRepo{
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
			return nil, nil
		},
	}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

func withResolveRepoMarket(market *markets.Market) func(*resolveRepo) {
	return func(repo *resolveRepo) {
		repo.getByIDFunc = func(context.Context, int64) (*markets.Market, error) {
			if market == nil {
				return nil, markets.ErrMarketNotFound
			}
			return market, nil
		}
	}
}

func withResolveRepoBets(bets []*markets.Bet) func(*resolveRepo) {
	return func(repo *resolveRepo) {
		repo.listBetsForMarketFunc = func(context.Context, int64) ([]*markets.Bet, error) {
			return bets, nil
		}
	}
}

func withResolveRepoPayouts(positions []*markets.PayoutPosition) func(*resolveRepo) {
	return func(repo *resolveRepo) {
		repo.calculatePayoutPositionsFunc = func(context.Context, int64) ([]*markets.PayoutPosition, error) {
			return positions, nil
		}
	}
}

func withResolveRepoResolve(fn func(context.Context, int64, string) error) func(*resolveRepo) {
	return func(repo *resolveRepo) {
		repo.resolveMarketFunc = fn
	}
}

func (r *resolveRepo) Create(ctx context.Context, market *markets.Market) error {
	if r.createFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.createFunc(ctx, market)
}
func (r *resolveRepo) UpdateLabels(ctx context.Context, id int64, yesLabel string, noLabel string) error {
	if r.updateLabelsFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.updateLabelsFunc(ctx, id, yesLabel, noLabel)
}
func (r *resolveRepo) List(ctx context.Context, filters markets.ListFilters) ([]*markets.Market, error) {
	if r.listFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listFunc(ctx, filters)
}
func (r *resolveRepo) ListByStatus(ctx context.Context, status string, page markets.Page) ([]*markets.Market, error) {
	if r.listByStatusFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listByStatusFunc(ctx, status, page)
}
func (r *resolveRepo) Search(ctx context.Context, query string, filters markets.SearchFilters) ([]*markets.Market, error) {
	if r.searchFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.searchFunc(ctx, query, filters)
}
func (r *resolveRepo) Delete(ctx context.Context, id int64) error {
	if r.deleteFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.deleteFunc(ctx, id)
}

func (r *resolveRepo) GetByID(ctx context.Context, id int64) (*markets.Market, error) {
	if r.getByIDFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.getByIDFunc(ctx, id)
}

func (r *resolveRepo) ResolveMarket(ctx context.Context, marketID int64, resolution string) error {
	if r.resolveMarketFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.resolveMarketFunc(ctx, marketID, resolution)
}

func (r *resolveRepo) GetUserPosition(ctx context.Context, marketID int64, username string) (*markets.UserPosition, error) {
	if r.getUserPositionFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.getUserPositionFunc(ctx, marketID, username)
}

func (r *resolveRepo) ListMarketPositions(ctx context.Context, marketID int64) (markets.MarketPositions, error) {
	if r.listMarketPositionsFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listMarketPositionsFunc(ctx, marketID)
}

func (r *resolveRepo) ListBetsForMarket(ctx context.Context, marketID int64) ([]*markets.Bet, error) {
	if r.listBetsForMarketFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listBetsForMarketFunc(ctx, marketID)
}

func (r *resolveRepo) CalculatePayoutPositions(ctx context.Context, marketID int64) ([]*markets.PayoutPosition, error) {
	if r.calculatePayoutPositionsFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.calculatePayoutPositionsFunc(ctx, marketID)
}

func (r *resolveRepo) GetPublicMarket(ctx context.Context, marketID int64) (*markets.PublicMarket, error) {
	if r.getPublicMarketFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.getPublicMarketFunc(ctx, marketID)
}

type resolveUserService struct {
	validateUserExistsFunc  func(context.Context, string) error
	validateUserBalanceFunc func(context.Context, string, int64, int64) error
	deductBalanceFunc       func(context.Context, string, int64) error
	applyTransactionFunc    func(context.Context, string, int64, string) error
	getPublicUserFunc       func(context.Context, string) (*users.PublicUser, error)
	applied                 []struct {
		username string
		amount   int64
		txType   string
	}
}

func newResolveUserService(opts ...func(*resolveUserService)) *resolveUserService {
	service := &resolveUserService{
		validateUserExistsFunc:  func(context.Context, string) error { return nil },
		validateUserBalanceFunc: func(context.Context, string, int64, int64) error { return nil },
		deductBalanceFunc:       func(context.Context, string, int64) error { return nil },
		applyTransactionFunc:    func(context.Context, string, int64, string) error { return nil },
		getPublicUserFunc:       func(context.Context, string) (*users.PublicUser, error) { return nil, nil },
	}
	for _, opt := range opts {
		opt(service)
	}
	return service
}

func (s *resolveUserService) ValidateUserExists(ctx context.Context, username string) error {
	if s.validateUserExistsFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return s.validateUserExistsFunc(ctx, username)
}
func (s *resolveUserService) ValidateUserBalance(ctx context.Context, username string, amount int64, maxDebt int64) error {
	if s.validateUserBalanceFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return s.validateUserBalanceFunc(ctx, username, amount, maxDebt)
}
func (s *resolveUserService) DeductBalance(ctx context.Context, username string, amount int64) error {
	if s.deductBalanceFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return s.deductBalanceFunc(ctx, username, amount)
}
func (s *resolveUserService) ApplyTransaction(ctx context.Context, username string, amount int64, tx string) error {
	if s.applyTransactionFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	if err := s.applyTransactionFunc(ctx, username, amount, tx); err != nil {
		return err
	}
	s.applied = append(s.applied, struct {
		username string
		amount   int64
		txType   string
	}{username: username, amount: amount, txType: tx})
	return nil
}

func (s *resolveUserService) GetPublicUser(ctx context.Context, username string) (*users.PublicUser, error) {
	if s.getPublicUserFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return s.getPublicUserFunc(ctx, username)
}

type nopClock struct{ nowFunc func() time.Time }

func newNopClock(now time.Time) nopClock {
	return nopClock{nowFunc: func() time.Time { return now }}
}

func (c nopClock) Now() time.Time {
	if c.nowFunc == nil {
		return marketsTestTime()
	}
	return c.nowFunc()
}

func TestResolveMarketRefundsOnNA(t *testing.T) {
	market := &markets.Market{
		ID:              1,
		CreatorUsername: "creator",
		Status:          "active",
	}
	repo := newResolveRepo(
		withResolveRepoMarket(market),
		withResolveRepoResolve(func(context.Context, int64, string) error {
			market.Status = "resolved"
			return nil
		}),
		withResolveRepoBets([]*markets.Bet{
			{Username: "alice", Amount: 50},
			{Username: "bob", Amount: 30},
		}),
	)
	userSvc := newResolveUserService()
	service := markets.NewService(repo, userSvc, newNopClock(marketsTestTime()), markets.Config{})

	if err := service.ResolveMarket(context.Background(), 1, "N/A", "creator"); err != nil {
		t.Fatalf("ResolveMarket returned error: %v", err)
	}

	if len(userSvc.applied) != 2 {
		t.Fatalf("expected 2 refund transactions, got %d", len(userSvc.applied))
	}

	for _, call := range userSvc.applied {
		if call.txType != users.TransactionRefund {
			t.Fatalf("expected refund transaction, got %s", call.txType)
		}
	}

	if got := (nopClock{}).Now(); !got.Equal(marketsTestTime()) {
		t.Fatalf("expected zero-value clock fallback, got %v", got)
	}
}

func TestResolveMarketPaysWinners(t *testing.T) {
	market := &markets.Market{
		ID:              42,
		CreatorUsername: "creator",
		Status:          "active",
	}
	repo := newResolveRepo(
		withResolveRepoMarket(market),
		withResolveRepoResolve(func(context.Context, int64, string) error {
			market.Status = "resolved"
			return nil
		}),
		withResolveRepoPayouts([]*markets.PayoutPosition{
			{Username: "winner", Value: 120},
			{Username: "loser", Value: 0},
		}),
	)
	userSvc := newResolveUserService()
	service := markets.NewService(repo, userSvc, newNopClock(marketsTestTime()), markets.Config{})

	if err := service.ResolveMarket(context.Background(), 42, "YES", "creator"); err != nil {
		t.Fatalf("ResolveMarket returned error: %v", err)
	}

	if len(userSvc.applied) != 1 {
		t.Fatalf("expected single payout, got %d", len(userSvc.applied))
	}

	positions, err := repo.CalculatePayoutPositions(context.Background(), 42)
	if err != nil {
		t.Fatalf("CalculatePayoutPositions returned error: %v", err)
	}
	if !positions[0].IsPayable() || positions[1].IsPayable() {
		t.Fatalf("unexpected payout classification: %+v", positions)
	}

	call := userSvc.applied[0]
	if call.username != "winner" || call.amount != 120 || call.txType != users.TransactionWin {
		t.Fatalf("unexpected payout %+v", call)
	}

	if _, err := (&resolveRepo{}).CalculatePayoutPositions(context.Background(), 1); !errors.Is(err, errUnexpectedMarketsTestCall) {
		t.Fatalf("expected zero-value repo to fail predictably, got %v", err)
	}
}

func TestResolveMarketRejectsUnauthorized(t *testing.T) {
	repo := newResolveRepo(withResolveRepoMarket(&markets.Market{
		ID:              5,
		CreatorUsername: "owner",
		Status:          "active",
	}))
	userSvc := newResolveUserService()
	service := markets.NewService(repo, userSvc, newNopClock(marketsTestTime()), markets.Config{})

	err := service.ResolveMarket(context.Background(), 5, "YES", "intruder")
	requireUnauthorized(t, err)

	if _, err := (&resolveUserService{}).GetPublicUser(context.Background(), "owner"); !errors.Is(err, errUnexpectedMarketsTestCall) {
		t.Fatalf("expected zero-value user service to fail predictably, got %v", err)
	}
}
