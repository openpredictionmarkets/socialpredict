package bets_test

import (
	"context"
	"errors"
	"testing"
	"time"

	bets "socialpredict/internal/domain/bets"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/setup"
)

var errUnexpectedServiceCall = errors.New("unexpected call")

type fakeRepo struct {
	created    *models.Bet
	createFunc func(ctx context.Context, bet *models.Bet) error
	hasBetFunc func(ctx context.Context, marketID uint, username string) (bool, error)
}

func newFakeRepo(opts ...func(*fakeRepo)) *fakeRepo {
	repo := &fakeRepo{
		createFunc: func(context.Context, *models.Bet) error { return nil },
		hasBetFunc: func(context.Context, uint, string) (bool, error) { return false, nil },
	}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

func withFakeRepoCreate(fn func(ctx context.Context, bet *models.Bet) error) func(*fakeRepo) {
	return func(repo *fakeRepo) {
		repo.createFunc = fn
	}
}

func withFakeRepoHasBet(fn func(ctx context.Context, marketID uint, username string) (bool, error)) func(*fakeRepo) {
	return func(repo *fakeRepo) {
		repo.hasBetFunc = fn
	}
}

func (f *fakeRepo) Create(ctx context.Context, bet *models.Bet) error {
	if err := f.createFunc(ctx, bet); err != nil {
		return err
	}
	copied := *bet
	f.created = &copied
	return nil
}

func (f *fakeRepo) UserHasBet(ctx context.Context, marketID uint, username string) (bool, error) {
	return f.hasBetFunc(ctx, marketID, username)
}

type fakeMarkets struct {
	getMarketFunc           func(ctx context.Context, id int64) (*dmarkets.Market, error)
	getUserPositionInMarket func(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error)
}

func newFakeMarkets(opts ...func(*fakeMarkets)) *fakeMarkets {
	markets := &fakeMarkets{
		getMarketFunc: func(context.Context, int64) (*dmarkets.Market, error) {
			return nil, errUnexpectedServiceCall
		},
		getUserPositionInMarket: func(context.Context, int64, string) (*dmarkets.UserPosition, error) {
			return nil, errUnexpectedServiceCall
		},
	}
	for _, opt := range opts {
		opt(markets)
	}
	return markets
}

func withFakeMarket(fn func(ctx context.Context, id int64) (*dmarkets.Market, error)) func(*fakeMarkets) {
	return func(markets *fakeMarkets) {
		markets.getMarketFunc = fn
	}
}

func withFakePosition(fn func(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error)) func(*fakeMarkets) {
	return func(markets *fakeMarkets) {
		markets.getUserPositionInMarket = fn
	}
}

func (f *fakeMarkets) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	return f.getMarketFunc(ctx, id)
}

func (f *fakeMarkets) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	return f.getUserPositionInMarket(ctx, marketID, username)
}

type applyCall struct {
	username    string
	amount      int64
	transaction string
}

type fakeUsers struct {
	calls                []applyCall
	getUserFunc          func(ctx context.Context, username string) (*dusers.User, error)
	applyTransactionFunc func(ctx context.Context, username string, amount int64, transactionType string) error
}

func newFakeUsers(opts ...func(*fakeUsers)) *fakeUsers {
	users := &fakeUsers{
		getUserFunc: func(context.Context, string) (*dusers.User, error) {
			return nil, errUnexpectedServiceCall
		},
		applyTransactionFunc: func(context.Context, string, int64, string) error { return nil },
	}
	for _, opt := range opts {
		opt(users)
	}
	return users
}

func withFakeUserLookup(fn func(ctx context.Context, username string) (*dusers.User, error)) func(*fakeUsers) {
	return func(users *fakeUsers) {
		users.getUserFunc = fn
	}
}

func withFakeApplyTransaction(fn func(ctx context.Context, username string, amount int64, transactionType string) error) func(*fakeUsers) {
	return func(users *fakeUsers) {
		users.applyTransactionFunc = fn
	}
}

func (f *fakeUsers) GetUser(ctx context.Context, username string) (*dusers.User, error) {
	return f.getUserFunc(ctx, username)
}

func (f *fakeUsers) ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error {
	if err := f.applyTransactionFunc(ctx, username, amount, transactionType); err != nil {
		return err
	}
	f.calls = append(f.calls, applyCall{username: username, amount: amount, transaction: transactionType})
	return nil
}

type fixedClock struct {
	nowFunc func() time.Time
}

func newFixedClock(now time.Time) fixedClock {
	return fixedClock{
		nowFunc: func() time.Time { return now },
	}
}

func (c fixedClock) Now() time.Time {
	if c.nowFunc == nil {
		return time.Time{}
	}
	return c.nowFunc()
}

type serviceFixture struct {
	econ    *setup.EconomicConfig
	repo    *fakeRepo
	markets *fakeMarkets
	users   *fakeUsers
	clock   fixedClock
}

type serviceFixtureOption func(*serviceFixture)

func withFixtureMarket(market *dmarkets.Market) serviceFixtureOption {
	return func(f *serviceFixture) {
		f.markets = newFakeMarkets(withFakeMarket(func(context.Context, int64) (*dmarkets.Market, error) {
			return market, nil
		}))
	}
}

func withFixturePosition(position *dmarkets.UserPosition) serviceFixtureOption {
	return func(f *serviceFixture) {
		if f.markets == nil {
			f.markets = newFakeMarkets()
		}
		current := f.markets.getMarketFunc
		f.markets = newFakeMarkets(
			withFakeMarket(current),
			withFakePosition(func(context.Context, int64, string) (*dmarkets.UserPosition, error) {
				return position, nil
			}),
		)
	}
}

func withFixtureUser(user *dusers.User) serviceFixtureOption {
	return func(f *serviceFixture) {
		f.users = newFakeUsers(withFakeUserLookup(func(context.Context, string) (*dusers.User, error) {
			return user, nil
		}))
	}
}

func withFixtureMaxDust(maxDust int64) serviceFixtureOption {
	return func(f *serviceFixture) {
		f.econ.Economics.Betting.MaxDustPerSale = maxDust
	}
}

func newServiceFixture(now time.Time, opts ...serviceFixtureOption) (*serviceFixture, *bets.Service) {
	econ := modelstesting.GenerateEconomicConfig()
	fixture := &serviceFixture{
		econ:    econ,
		repo:    newFakeRepo(),
		markets: newFakeMarkets(),
		users:   newFakeUsers(),
		clock:   newFixedClock(now),
	}
	for _, opt := range opts {
		opt(fixture)
	}
	svc := bets.NewService(fixture.repo, fixture.markets, fixture.users, fixture.econ, fixture.clock)
	return fixture, svc
}

func serviceTestTime() time.Time {
	return time.Date(2026, time.February, 3, 4, 5, 6, 0, time.UTC)
}

func TestServicePlace_Succeeds(t *testing.T) {
	now := serviceTestTime()
	fixture, svc := newServiceFixture(
		now,
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixtureUser(&dusers.User{Username: "alice", AccountBalance: 500}),
	)

	placed, err := svc.Place(context.Background(), bets.PlaceRequest{Username: "alice", MarketID: 1, Amount: 100, Outcome: "yes"})
	if err != nil {
		t.Fatalf("Place returned error: %v", err)
	}

	if placed.Username != "alice" || placed.Amount != 100 || placed.MarketID != 1 {
		t.Fatalf("unexpected placed bet: %+v", placed)
	}
	if !placed.PlacedAt.Equal(now) {
		t.Fatalf("expected placed time %v, got %v", now, placed.PlacedAt)
	}

	if fixture.repo.created == nil {
		t.Fatalf("expected repository Create to be called")
	}
	if fixture.repo.created.Outcome != "YES" {
		t.Fatalf("expected outcome YES, got %s", fixture.repo.created.Outcome)
	}

	if len(fixture.users.calls) != 1 {
		t.Fatalf("expected one ApplyTransaction call, got %d", len(fixture.users.calls))
	}
	totalCost := int64(100 + fixture.econ.Economics.Betting.BetFees.InitialBetFee + fixture.econ.Economics.Betting.BetFees.BuySharesFee)
	if fixture.users.calls[0].amount != totalCost {
		t.Fatalf("unexpected transaction amount: %d", fixture.users.calls[0].amount)
	}
}

func TestServicePlace_InsufficientBalance(t *testing.T) {
	now := serviceTestTime()
	_, svc := newServiceFixture(
		now,
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixtureUser(&dusers.User{Username: "alice", AccountBalance: 0}),
	)

	_, err := svc.Place(context.Background(), bets.PlaceRequest{Username: "alice", MarketID: 1, Amount: 9999, Outcome: "YES"})
	if !errors.Is(err, bets.ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
}

func TestServicePlace_InvalidOutcome(t *testing.T) {
	now := serviceTestTime()
	_, svc := newServiceFixture(
		now,
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixtureUser(&dusers.User{Username: "alice", AccountBalance: 100}),
	)

	_, err := svc.Place(context.Background(), bets.PlaceRequest{Username: "alice", MarketID: 1, Amount: 10, Outcome: "MAYBE"})
	if !errors.Is(err, bets.ErrInvalidOutcome) {
		t.Fatalf("expected ErrInvalidOutcome, got %v", err)
	}
}

func TestServicePlace_MarketClosed(t *testing.T) {
	now := serviceTestTime()
	_, svc := newServiceFixture(
		now,
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "resolved", ResolutionDateTime: now.Add(-time.Hour)}),
		withFixtureUser(&dusers.User{Username: "alice", AccountBalance: 100}),
	)

	_, err := svc.Place(context.Background(), bets.PlaceRequest{Username: "alice", MarketID: 1, Amount: 10, Outcome: "YES"})
	if !errors.Is(err, bets.ErrMarketClosed) {
		t.Fatalf("expected ErrMarketClosed, got %v", err)
	}
}

func TestServiceSell_Succeeds(t *testing.T) {
	now := serviceTestTime()
	fixture, svc := newServiceFixture(
		now,
		withFixtureMaxDust(0),
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixturePosition(&dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 10, NoSharesOwned: 0, Value: 100}),
		withFixtureUser(&dusers.User{Username: "alice"}),
	)

	res, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 25, Outcome: "YES"})
	if err != nil {
		t.Fatalf("Sell returned error: %v", err)
	}
	if res.SharesSold != 2 || res.SaleValue != 20 || res.Dust != 5 {
		t.Fatalf("unexpected sell result: %+v", res)
	}
	if !res.TransactionAt.Equal(now) {
		t.Fatalf("expected transaction time %v, got %v", now, res.TransactionAt)
	}
	if fixture.repo.created == nil || fixture.repo.created.Amount != -2 || fixture.repo.created.Outcome != "YES" {
		t.Fatalf("unexpected stored bet: %+v", fixture.repo.created)
	}
	if len(fixture.users.calls) != 1 || fixture.users.calls[0].transaction != dusers.TransactionSale || fixture.users.calls[0].amount != 20 {
		t.Fatalf("unexpected user transaction: %+v", fixture.users.calls)
	}
}

func TestServiceSell_NoPosition(t *testing.T) {
	now := serviceTestTime()
	_, svc := newServiceFixture(
		now,
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixturePosition(&dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 0, NoSharesOwned: 0, Value: 0}),
		withFixtureUser(&dusers.User{Username: "alice"}),
	)

	_, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 10, Outcome: "YES"})
	if !errors.Is(err, bets.ErrNoPosition) {
		t.Fatalf("expected ErrNoPosition, got %v", err)
	}
}

func TestServiceSell_DustCapExceeded(t *testing.T) {
	now := serviceTestTime()
	_, svc := newServiceFixture(
		now,
		withFixtureMaxDust(2),
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixturePosition(&dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 10, Value: 100}),
		withFixtureUser(&dusers.User{Username: "alice"}),
	)

	_, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 33, Outcome: "YES"})
	var dustErr bets.ErrDustCapExceeded
	if !errors.As(err, &dustErr) {
		t.Fatalf("expected ErrDustCapExceeded, got %v", err)
	}
}

func TestServiceSell_RequestTooSmall(t *testing.T) {
	now := serviceTestTime()
	_, svc := newServiceFixture(
		now,
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixturePosition(&dmarkets.UserPosition{
			Username:       "alice",
			MarketID:       1,
			YesSharesOwned: 5,
			Value:          50,
		}),
		withFixtureUser(&dusers.User{Username: "alice"}),
	)

	_, err := svc.Sell(context.Background(), bets.SellRequest{
		Username: "alice",
		MarketID: 1,
		Amount:   5, // less than value per share (10)
		Outcome:  "YES",
	})
	if !errors.Is(err, bets.ErrInvalidAmount) {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}
}
