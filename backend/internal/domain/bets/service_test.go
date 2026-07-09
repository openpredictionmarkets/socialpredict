package bets_test

import (
	"context"
	"errors"
	"testing"
	"time"

	bets "socialpredict/internal/domain/bets"
	"socialpredict/internal/domain/boundary"
	dmarkets "socialpredict/internal/domain/markets"
	positionsmath "socialpredict/internal/domain/math/positions"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/models/modelstesting"
)

// Bets service tests use fakes to keep accounting rules package-local and fast.
// They prove ordering and collaborator boundaries, not database truth. WAVE07
// added repository-level Postgres proof for place-bet transaction behavior.
// Sell-position settlement now has a repository transaction seam; broader
// market-resolution accounting remains separate.

var errUnexpectedServiceCall = errors.New("unexpected call")

type fakeRepo struct {
	created *boundary.Bet
	writer  fakeBetWriter
	history fakeBetHistoryReader
}

type fakePlaceUnit struct {
	repo  bets.Repository
	users bets.UserService
}

func (f fakePlaceUnit) PlaceBetTransaction(ctx context.Context, fn bets.PlaceTransactionFunc) error {
	return fn(ctx, f.repo, f.users)
}

type fakeSellUnit struct {
	repo    bets.Repository
	markets bets.MarketService
	users   bets.UserService
}

func (f fakeSellUnit) SellBetTransaction(ctx context.Context, fn bets.SellTransactionFunc) error {
	return fn(ctx, f.repo, f.markets, f.users)
}

func newFakeRepo(opts ...func(*fakeRepo)) *fakeRepo {
	repo := &fakeRepo{
		writer: fakeBetWriter{
			createFunc: func(context.Context, *boundary.Bet) error { return nil },
		},
		history: fakeBetHistoryReader{
			hasBetFunc: func(context.Context, uint, string) (bool, error) { return false, nil },
		},
	}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

func withFakeRepoCreate(fn func(ctx context.Context, bet *boundary.Bet) error) func(*fakeRepo) {
	return func(repo *fakeRepo) {
		repo.writer.createFunc = fn
	}
}

func withFakeRepoHasBet(fn func(ctx context.Context, marketID uint, username string) (bool, error)) func(*fakeRepo) {
	return func(repo *fakeRepo) {
		repo.history.hasBetFunc = fn
	}
}

type fakeBetWriter struct {
	createFunc func(ctx context.Context, bet *boundary.Bet) error
}

func (f *fakeRepo) Create(ctx context.Context, bet *boundary.Bet) error {
	return f.writer.Create(ctx, bet, f)
}

type fakeBetHistoryReader struct {
	hasBetFunc func(ctx context.Context, marketID uint, username string) (bool, error)
}

func (f *fakeRepo) UserHasBet(ctx context.Context, marketID uint, username string) (bool, error) {
	return f.history.UserHasBet(ctx, marketID, username)
}

func (f fakeBetWriter) Create(ctx context.Context, bet *boundary.Bet, repo *fakeRepo) error {
	if f.createFunc == nil {
		return errUnexpectedServiceCall
	}
	if err := f.createFunc(ctx, bet); err != nil {
		return err
	}
	if bet == nil {
		repo.created = nil
		return nil
	}
	copied := *bet
	repo.created = &copied
	return nil
}

func (f fakeBetHistoryReader) UserHasBet(ctx context.Context, marketID uint, username string) (bool, error) {
	if f.hasBetFunc == nil {
		return false, errUnexpectedServiceCall
	}
	return f.hasBetFunc(ctx, marketID, username)
}

type fakeMarkets struct {
	reader    fakeMarketReader
	positions fakePositionReader
	projector fakePositionProjector
}

func newFakeMarkets(opts ...func(*fakeMarkets)) *fakeMarkets {
	markets := &fakeMarkets{
		reader: fakeMarketReader{
			getMarketFunc: func(context.Context, int64) (*dmarkets.Market, error) {
				return nil, errUnexpectedServiceCall
			},
		},
		positions: fakePositionReader{
			getUserPositionInMarketFunc: func(context.Context, int64, string) (*dmarkets.UserPosition, error) {
				return nil, errUnexpectedServiceCall
			},
			getUserSellablePositionInMarketFunc: func(context.Context, int64, string, string) (*dmarkets.UserPosition, error) {
				return nil, errUnexpectedServiceCall
			},
		},
		projector: fakePositionProjector{
			projectUserPositionAfterBetFunc: func(context.Context, int64, string, boundary.Bet) (*dmarkets.UserPosition, error) {
				return nil, errUnexpectedServiceCall
			},
		},
	}
	for _, opt := range opts {
		opt(markets)
	}
	return markets
}

func withFakeMarket(fn func(ctx context.Context, id int64) (*dmarkets.Market, error)) func(*fakeMarkets) {
	return func(markets *fakeMarkets) {
		markets.reader.getMarketFunc = fn
	}
}

func withFakePosition(fn func(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error)) func(*fakeMarkets) {
	return func(markets *fakeMarkets) {
		markets.positions.getUserPositionInMarketFunc = fn
		markets.positions.getUserSellablePositionInMarketFunc = func(ctx context.Context, marketID int64, username string, _ string) (*dmarkets.UserPosition, error) {
			return fn(ctx, marketID, username)
		}
	}
}

func withFakeSellablePosition(fn func(ctx context.Context, marketID int64, username string, outcome string) (*dmarkets.UserPosition, error)) func(*fakeMarkets) {
	return func(markets *fakeMarkets) {
		markets.positions.getUserSellablePositionInMarketFunc = fn
	}
}

func withFakePositionProjection(fn func(ctx context.Context, marketID int64, username string, bet boundary.Bet) (*dmarkets.UserPosition, error)) func(*fakeMarkets) {
	return func(markets *fakeMarkets) {
		markets.projector.projectUserPositionAfterBetFunc = fn
	}
}

type fakeMarketReader struct {
	getMarketFunc func(ctx context.Context, id int64) (*dmarkets.Market, error)
}

func (f *fakeMarkets) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	return f.reader.GetMarket(ctx, id)
}

type fakePositionReader struct {
	getUserPositionInMarketFunc         func(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error)
	getUserSellablePositionInMarketFunc func(ctx context.Context, marketID int64, username string, outcome string) (*dmarkets.UserPosition, error)
}

func (f *fakeMarkets) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	return f.positions.GetUserPositionInMarket(ctx, marketID, username)
}

func (f *fakeMarkets) GetUserSellablePositionInMarket(ctx context.Context, marketID int64, username string, outcome string) (*dmarkets.UserPosition, error) {
	return f.positions.GetUserSellablePositionInMarket(ctx, marketID, username, outcome)
}

type fakePositionProjector struct {
	projectUserPositionAfterBetFunc func(ctx context.Context, marketID int64, username string, bet boundary.Bet) (*dmarkets.UserPosition, error)
}

func (f *fakeMarkets) ProjectUserPositionAfterBet(ctx context.Context, marketID int64, username string, bet boundary.Bet) (*dmarkets.UserPosition, error) {
	return f.projector.ProjectUserPositionAfterBet(ctx, marketID, username, bet)
}

func (f fakeMarketReader) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	if f.getMarketFunc == nil {
		return nil, errUnexpectedServiceCall
	}
	return f.getMarketFunc(ctx, id)
}

func (f fakePositionReader) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	if f.getUserPositionInMarketFunc == nil {
		return nil, errUnexpectedServiceCall
	}
	return f.getUserPositionInMarketFunc(ctx, marketID, username)
}

func (f fakePositionReader) GetUserSellablePositionInMarket(ctx context.Context, marketID int64, username string, outcome string) (*dmarkets.UserPosition, error) {
	if f.getUserSellablePositionInMarketFunc == nil {
		return nil, errUnexpectedServiceCall
	}
	return f.getUserSellablePositionInMarketFunc(ctx, marketID, username, outcome)
}

func (f fakePositionProjector) ProjectUserPositionAfterBet(ctx context.Context, marketID int64, username string, bet boundary.Bet) (*dmarkets.UserPosition, error) {
	if f.projectUserPositionAfterBetFunc == nil {
		return nil, errUnexpectedServiceCall
	}
	return f.projectUserPositionAfterBetFunc(ctx, marketID, username, bet)
}

type applyCall struct {
	username    string
	amount      int64
	transaction string
}

type fakeUsers struct {
	calls  []applyCall
	reader fakeUserReader
	writer fakeTransactionRecorder
}

func newFakeUsers(opts ...func(*fakeUsers)) *fakeUsers {
	users := &fakeUsers{
		reader: fakeUserReader{
			getUserFunc: func(context.Context, string) (*dusers.User, error) {
				return nil, errUnexpectedServiceCall
			},
		},
		writer: fakeTransactionRecorder{
			applyTransactionFunc: func(context.Context, string, int64, string) error { return nil },
		},
	}
	for _, opt := range opts {
		opt(users)
	}
	return users
}

func withFakeUserLookup(fn func(ctx context.Context, username string) (*dusers.User, error)) func(*fakeUsers) {
	return func(users *fakeUsers) {
		users.reader.getUserFunc = fn
	}
}

func withFakeApplyTransaction(fn func(ctx context.Context, username string, amount int64, transactionType string) error) func(*fakeUsers) {
	return func(users *fakeUsers) {
		users.writer.applyTransactionFunc = fn
	}
}

type fakeUserReader struct {
	getUserFunc func(ctx context.Context, username string) (*dusers.User, error)
}

func (f *fakeUsers) GetUser(ctx context.Context, username string) (*dusers.User, error) {
	return f.reader.GetUser(ctx, username)
}

type fakeTransactionRecorder struct {
	applyTransactionFunc func(ctx context.Context, username string, amount int64, transactionType string) error
}

func (f *fakeUsers) ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error {
	return f.writer.ApplyTransaction(ctx, username, amount, transactionType, &f.calls)
}

func (f fakeUserReader) GetUser(ctx context.Context, username string) (*dusers.User, error) {
	if f.getUserFunc == nil {
		return nil, errUnexpectedServiceCall
	}
	return f.getUserFunc(ctx, username)
}

func (f fakeTransactionRecorder) ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string, calls *[]applyCall) error {
	if f.applyTransactionFunc == nil {
		return errUnexpectedServiceCall
	}
	if err := f.applyTransactionFunc(ctx, username, amount, transactionType); err != nil {
		return err
	}
	*calls = append(*calls, applyCall{username: username, amount: amount, transaction: transactionType})
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
		return serviceTestTime()
	}
	return c.nowFunc()
}

type serviceFixture struct {
	config  bets.Config
	repo    *fakeRepo
	markets *fakeMarkets
	users   *fakeUsers
	clock   fixedClock
}

type serviceFixtureOption func(*serviceFixture)

func withFixtureMarket(market *dmarkets.Market) serviceFixtureOption {
	return func(f *serviceFixture) {
		if f.markets == nil {
			f.markets = newFakeMarkets()
		}
		f.markets.reader.getMarketFunc = func(context.Context, int64) (*dmarkets.Market, error) {
			return market, nil
		}
	}
}

func withFixturePosition(position *dmarkets.UserPosition) serviceFixtureOption {
	return func(f *serviceFixture) {
		if f.markets == nil {
			f.markets = newFakeMarkets()
		}
		f.markets.positions.getUserPositionInMarketFunc = func(context.Context, int64, string) (*dmarkets.UserPosition, error) {
			return position, nil
		}
		f.markets.positions.getUserSellablePositionInMarketFunc = func(context.Context, int64, string, string) (*dmarkets.UserPosition, error) {
			return position, nil
		}
		f.markets.projector.projectUserPositionAfterBetFunc = func(_ context.Context, _ int64, _ string, bet boundary.Bet) (*dmarkets.UserPosition, error) {
			return projectedFixturePosition(position, bet), nil
		}
	}
}

func withFixtureProjection(projected *dmarkets.UserPosition) serviceFixtureOption {
	return func(f *serviceFixture) {
		if f.markets == nil {
			f.markets = newFakeMarkets()
		}
		f.markets.projector.projectUserPositionAfterBetFunc = func(context.Context, int64, string, boundary.Bet) (*dmarkets.UserPosition, error) {
			return projected, nil
		}
	}
}

func projectedFixturePosition(position *dmarkets.UserPosition, bet boundary.Bet) *dmarkets.UserPosition {
	if position == nil {
		return nil
	}
	projected := *position
	if bet.Amount >= 0 {
		return &projected
	}

	sharesSold := -bet.Amount
	switch bet.Outcome {
	case "YES":
		projected.YesSharesOwned -= sharesSold
		if projected.YesSharesOwned < 0 {
			projected.YesSharesOwned = 0
		}
		projected.Value -= sharesSold * fixtureValuePerShare(position.Value, position.YesSharesOwned)
	case "NO":
		projected.NoSharesOwned -= sharesSold
		if projected.NoSharesOwned < 0 {
			projected.NoSharesOwned = 0
		}
		projected.Value -= sharesSold * fixtureValuePerShare(position.Value, position.NoSharesOwned)
	}
	if projected.Value < 0 {
		projected.Value = 0
	}
	return &projected
}

func fixtureValuePerShare(value, shares int64) int64 {
	if value <= 0 || shares <= 0 {
		return 0
	}
	return value / shares
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
		f.config.MaxDustPerSale = maxDust
	}
}

func defaultBetsConfig() bets.Config {
	econ := modelstesting.GenerateEconomicConfig()
	return bets.Config{
		InitialBetFee:      econ.Economics.Betting.BetFees.InitialBetFee,
		BuySharesFee:       econ.Economics.Betting.BetFees.BuySharesFee,
		MaxDustPerSale:     econ.Economics.Betting.MaxDustPerSale,
		MaximumDebtAllowed: econ.Economics.User.MaximumDebtAllowed,
	}
}

func newServiceFixture(now time.Time, opts ...serviceFixtureOption) (*serviceFixture, *bets.Service) {
	fixture := &serviceFixture{
		config:  defaultBetsConfig(),
		repo:    newFakeRepo(),
		markets: newFakeMarkets(),
		users:   newFakeUsers(),
		clock:   newFixedClock(now),
	}
	for _, opt := range opts {
		opt(fixture)
	}
	placeUnit := fakePlaceUnit{repo: fixture.repo, users: fixture.users}
	sellUnit := fakeSellUnit{repo: fixture.repo, markets: fixture.markets, users: fixture.users}
	svc := bets.NewService(
		fixture.repo,
		fixture.markets,
		fixture.users,
		fixture.config,
		fixture.clock,
		bets.WithPlaceUnitOfWork(placeUnit),
		bets.WithSellUnitOfWork(sellUnit),
	)
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
	totalCost := int64(100 + fixture.config.InitialBetFee + fixture.config.BuySharesFee)
	if fixture.users.calls[0].amount != totalCost {
		t.Fatalf("unexpected transaction amount: %d", fixture.users.calls[0].amount)
	}

	fallbackService := bets.NewService(fixture.repo, fixture.markets, fixture.users, bets.Config{}, nil, bets.WithClock(nil))
	if fallbackService == nil {
		t.Fatalf("expected fallback service")
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

	if _, err := (&fakeRepo{}).UserHasBet(context.Background(), 1, "alice"); !errors.Is(err, errUnexpectedServiceCall) {
		t.Fatalf("expected zero-value repo to fail predictably, got %v", err)
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

	if _, err := (&fakeUsers{}).GetUser(context.Background(), "alice"); !errors.Is(err, errUnexpectedServiceCall) {
		t.Fatalf("expected zero-value users to fail predictably, got %v", err)
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

	if _, err := (&fakeMarkets{}).GetMarket(context.Background(), 1); !errors.Is(err, errUnexpectedServiceCall) {
		t.Fatalf("expected zero-value markets to fail predictably, got %v", err)
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
	if res.SharesSold != 2 || res.SaleValue != 20 || res.Dust != 0 || res.NetProceeds != 20 {
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

	if got := (fixedClock{}).Now(); got != serviceTestTime() {
		t.Fatalf("expected zero-value clock fallback, got %v", got)
	}
}

func TestServiceSell_DustAtCapCreditsNetProceeds(t *testing.T) {
	now := serviceTestTime()
	fixture, svc := newServiceFixture(
		now,
		withFixtureMaxDust(2),
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixturePosition(&dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 10, Value: 100}),
		withFixtureUser(&dusers.User{Username: "alice"}),
	)

	res, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 32, Outcome: "YES"})
	if err != nil {
		t.Fatalf("Sell returned error: %v", err)
	}
	if res.SharesSold != 3 || res.SaleValue != 30 || res.Dust != 2 || res.NetProceeds != 28 {
		t.Fatalf("unexpected sell result: %+v", res)
	}
	if fixture.repo.created == nil || fixture.repo.created.Amount != -3 {
		t.Fatalf("unexpected stored sale bet: %+v", fixture.repo.created)
	}
	if len(fixture.users.calls) != 1 || fixture.users.calls[0].transaction != dusers.TransactionSale || fixture.users.calls[0].amount != 28 {
		t.Fatalf("unexpected user transaction: %+v", fixture.users.calls)
	}
}

func TestServiceSell_OneCreditShareDoesNotCashOutHistoricalVolume(t *testing.T) {
	now := serviceTestTime()
	fixture, svc := newServiceFixture(
		now,
		withFixtureMaxDust(0),
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixturePosition(&dmarkets.UserPosition{Username: "alice", MarketID: 1, NoSharesOwned: 1, Value: 1}),
		withFixtureUser(&dusers.User{Username: "alice"}),
	)

	quote, err := svc.QuoteSell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 141, Outcome: "NO"})
	if err != nil {
		t.Fatalf("QuoteSell returned error: %v", err)
	}
	if quote.SharesSold != 1 || quote.SaleValue != 1 || quote.NetProceeds != 1 || quote.RequestedCredits != 1 {
		t.Fatalf("quote cashed out more than the current share value: %+v", quote)
	}

	result, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 141, Outcome: "NO"})
	if err != nil {
		t.Fatalf("Sell returned error: %v", err)
	}
	if result.SharesSold != 1 || result.SaleValue != 1 || result.NetProceeds != 1 {
		t.Fatalf("sell cashed out more than the current share value: %+v", result)
	}
	if len(fixture.users.calls) != 1 || fixture.users.calls[0].amount != 1 {
		t.Fatalf("unexpected user credit: %+v", fixture.users.calls)
	}
}

func TestServiceSell_FullPositionSellRequiresZeroProjectedValue(t *testing.T) {
	now := serviceTestTime()

	t.Run("full position sell succeeds at zero projected value", func(t *testing.T) {
		fixture, svc := newServiceFixture(
			now,
			withFixtureMaxDust(0),
			withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
			withFixturePosition(&dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 10, Value: 100}),
			withFixtureProjection(&dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 0, Value: 0}),
			withFixtureUser(&dusers.User{Username: "alice"}),
		)

		result, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 100, Outcome: "YES"})
		if err != nil {
			t.Fatalf("Sell returned error: %v", err)
		}
		if result.SharesSold != 10 || result.SaleValue != 100 || result.NetProceeds != 100 {
			t.Fatalf("unexpected full-position sell result: %+v", result)
		}
		if fixture.repo.created == nil || fixture.repo.created.Amount != -10 {
			t.Fatalf("unexpected sale ledger row: %+v", fixture.repo.created)
		}
		if len(fixture.users.calls) != 1 || fixture.users.calls[0].amount != 100 {
			t.Fatalf("unexpected user credit: %+v", fixture.users.calls)
		}
	})

	t.Run("full position sell rejects residual projected value", func(t *testing.T) {
		fixture, svc := newServiceFixture(
			now,
			withFixtureMaxDust(0),
			withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
			withFixturePosition(&dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 10, Value: 100}),
			withFixtureProjection(&dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 0, Value: 1}),
			withFixtureUser(&dusers.User{Username: "alice"}),
		)

		_, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 100, Outcome: "YES"})
		if !errors.Is(err, bets.ErrInsufficientShares) {
			t.Fatalf("expected residual projected value to return ErrInsufficientShares, got %v", err)
		}
		if fixture.repo.created != nil || len(fixture.users.calls) != 0 {
			t.Fatalf("residual projected value must not mutate ledger: repo=%+v users=%+v", fixture.repo.created, fixture.users.calls)
		}
	})
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

	if _, err := (&fakeMarkets{}).GetUserPositionInMarket(context.Background(), 1, "alice"); !errors.Is(err, errUnexpectedServiceCall) {
		t.Fatalf("expected zero-value position lookup to fail predictably, got %v", err)
	}
}

func TestServiceSell_NoSellableShares(t *testing.T) {
	now := serviceTestTime()

	t.Run("quote rejects aggregate shares with no unlocked value", func(t *testing.T) {
		_, svc := newServiceFixture(
			now,
			withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
			withFixturePosition(&dmarkets.UserPosition{Username: "alice", MarketID: 1, NoSharesOwned: 1, Value: 1}),
			func(f *serviceFixture) {
				f.markets.positions.getUserSellablePositionInMarketFunc = func(context.Context, int64, string, string) (*dmarkets.UserPosition, error) {
					return &dmarkets.UserPosition{Username: "alice", MarketID: 1, NoSharesOwned: 0, Value: 0}, nil
				}
			},
			withFixtureUser(&dusers.User{Username: "alice"}),
		)

		_, err := svc.QuoteSell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 1, Outcome: "NO"})
		if !errors.Is(err, bets.ErrNoSellableShares) {
			t.Fatalf("expected ErrNoSellableShares, got %v", err)
		}
	})

	t.Run("sell rejects before ledger mutation", func(t *testing.T) {
		fixture, svc := newServiceFixture(
			now,
			withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
			withFixturePosition(&dmarkets.UserPosition{Username: "alice", MarketID: 1, NoSharesOwned: 1, Value: 1}),
			func(f *serviceFixture) {
				f.markets.positions.getUserSellablePositionInMarketFunc = func(context.Context, int64, string, string) (*dmarkets.UserPosition, error) {
					return &dmarkets.UserPosition{Username: "alice", MarketID: 1, NoSharesOwned: 0, Value: 0}, nil
				}
			},
			withFixtureUser(&dusers.User{Username: "alice"}),
		)

		_, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 1, Outcome: "NO"})
		if !errors.Is(err, bets.ErrNoSellableShares) {
			t.Fatalf("expected ErrNoSellableShares, got %v", err)
		}
		if fixture.repo.created != nil || len(fixture.users.calls) != 0 {
			t.Fatalf("locked sell must not mutate ledger: repo=%+v users=%+v", fixture.repo.created, fixture.users.calls)
		}
	})
}

func TestServiceSell_UsesUnlockedSellableCap(t *testing.T) {
	now := serviceTestTime()
	fixture, svc := newServiceFixture(
		now,
		withFixtureMaxDust(1),
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixturePosition(&dmarkets.UserPosition{Username: "alice", MarketID: 1, NoSharesOwned: 3, Value: 3}),
		func(f *serviceFixture) {
			f.markets.positions.getUserSellablePositionInMarketFunc = func(context.Context, int64, string, string) (*dmarkets.UserPosition, error) {
				return &dmarkets.UserPosition{Username: "alice", MarketID: 1, NoSharesOwned: 2, Value: 2}, nil
			}
		},
		withFixtureUser(&dusers.User{Username: "alice"}),
	)

	quote, err := svc.QuoteSell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 3, Outcome: "NO"})
	if err != nil {
		t.Fatalf("QuoteSell returned error: %v", err)
	}
	if quote.SharesSold != 2 || quote.SaleValue != 2 || quote.Dust != 1 || quote.NetProceeds != 1 {
		t.Fatalf("quote did not respect sellable cap: %+v", quote)
	}

	result, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 3, Outcome: "NO"})
	if err != nil {
		t.Fatalf("Sell returned error: %v", err)
	}
	if result.SharesSold != 2 || result.SaleValue != 2 || result.Dust != 1 || result.NetProceeds != 1 {
		t.Fatalf("sell did not respect sellable cap: %+v", result)
	}
	if fixture.repo.created == nil || fixture.repo.created.Amount != -2 {
		t.Fatalf("unexpected sale row: %+v", fixture.repo.created)
	}
	if len(fixture.users.calls) != 1 || fixture.users.calls[0].amount != 1 {
		t.Fatalf("unexpected sale credit: %+v", fixture.users.calls)
	}
}

func TestServiceSell_RoundsDustDownToCap(t *testing.T) {
	now := serviceTestTime()
	fixture, svc := newServiceFixture(
		now,
		withFixtureMaxDust(2),
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixturePosition(&dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 10, Value: 100}),
		withFixtureUser(&dusers.User{Username: "alice"}),
	)

	result, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 33, Outcome: "YES"})
	if err != nil {
		t.Fatalf("Sell returned error: %v", err)
	}
	if result.SharesSold != 3 || result.SaleValue != 30 || result.Dust != 2 || result.NetProceeds != 28 {
		t.Fatalf("unexpected rounded sale result: %+v", result)
	}
	if fixture.repo.created == nil || fixture.repo.created.Amount != -3 {
		t.Fatalf("expected stored sale bet, got %+v", fixture.repo.created)
	}
}

func TestServiceQuoteSell_AllowsDustAtCapWithoutMutatingState(t *testing.T) {
	now := serviceTestTime()
	fixture, svc := newServiceFixture(
		now,
		withFixtureMaxDust(2),
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixturePosition(&dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 10, Value: 100}),
		withFixtureUser(&dusers.User{Username: "alice"}),
	)

	quote, err := svc.QuoteSell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 32, Outcome: "YES"})
	if err != nil {
		t.Fatalf("QuoteSell returned error: %v", err)
	}
	if !quote.Allowed || quote.SaleValue != 30 || quote.Dust != 2 || quote.NetProceeds != 28 || quote.MaxDust != 2 || quote.ValuePerShare != 10 {
		t.Fatalf("unexpected quote: %+v", quote)
	}
	if quote.DustCapCoverage != 0.3 {
		t.Fatalf("expected dust cap coverage 0.3, got %v", quote.DustCapCoverage)
	}
	if !containsInt64(quote.SuggestedAmounts, 32) {
		t.Fatalf("expected valid requested amount suggestions to include current request, got %+v", quote.SuggestedAmounts)
	}
	if fixture.repo.created != nil || len(fixture.users.calls) != 0 {
		t.Fatalf("quote should not mutate repo or user ledger: repo=%+v users=%+v", fixture.repo.created, fixture.users.calls)
	}
}

func TestServiceQuoteSell_OverCapRoundsPreviewToCap(t *testing.T) {
	now := serviceTestTime()
	fixture, svc := newServiceFixture(
		now,
		withFixtureMaxDust(2),
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixturePosition(&dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 10, Value: 100}),
		withFixtureUser(&dusers.User{Username: "alice"}),
	)

	quote, err := svc.QuoteSell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 33, Outcome: "YES"})
	if err != nil {
		t.Fatalf("QuoteSell returned error: %v", err)
	}
	if !quote.Allowed || quote.DustCapExceeded || quote.DustCapExceededBy != 0 {
		t.Fatalf("expected rounded allowed quote, got %+v", quote)
	}
	if quote.SaleValue != 30 || quote.Dust != 2 || quote.NetProceeds != 28 || quote.MaxDust != 2 {
		t.Fatalf("unexpected over-cap quote amounts: %+v", quote)
	}
	for _, want := range []int64{30, 31, 32} {
		if !containsInt64(quote.SuggestedAmounts, want) {
			t.Fatalf("expected suggestions to include %d, got %+v", want, quote.SuggestedAmounts)
		}
	}
	if fixture.repo.created != nil || len(fixture.users.calls) != 0 {
		t.Fatalf("quote should not mutate repo or user ledger: repo=%+v users=%+v", fixture.repo.created, fixture.users.calls)
	}
}

func TestServiceQuoteSell_RejectsUnsafeProjectedSale(t *testing.T) {
	now := serviceTestTime()
	_, svc := newServiceFixture(
		now,
		withFixtureMaxDust(1),
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixturePosition(&dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 2, Value: 502}),
		withFixtureProjection(&dmarkets.UserPosition{Username: "alice", MarketID: 1, NoSharesOwned: 2, Value: 500}),
		withFixtureUser(&dusers.User{Username: "alice"}),
	)

	_, err := svc.QuoteSell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 502, Outcome: "YES"})
	if !errors.Is(err, bets.ErrInsufficientShares) {
		t.Fatalf("expected unsafe projected quote to return ErrInsufficientShares, got %v", err)
	}
}

func TestServiceSell_RejectsUnsafeProjectedSaleBeforeMutatingLedger(t *testing.T) {
	now := serviceTestTime()
	current := &dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 2, Value: 502}
	fixture, svc := newServiceFixture(
		now,
		withFixtureMaxDust(1),
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixturePosition(current),
		withFixtureProjection(&dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 0, NoSharesOwned: 2, Value: 500}),
		withFixtureUser(&dusers.User{Username: "alice"}),
	)

	_, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 502, Outcome: "YES"})
	if !errors.Is(err, bets.ErrInsufficientShares) {
		t.Fatalf("expected unsafe projected sale to return ErrInsufficientShares, got %v", err)
	}
	if fixture.repo.created != nil {
		t.Fatalf("unsafe sale must not create a ledger row: %+v", fixture.repo.created)
	}
	if len(fixture.users.calls) != 0 {
		t.Fatalf("unsafe sale must not credit user ledger: %+v", fixture.users.calls)
	}
}

func TestServiceSell_RejectsProjectedSaleThatKeepsTooMuchValue(t *testing.T) {
	now := serviceTestTime()
	fixture, svc := newServiceFixture(
		now,
		withFixtureMaxDust(0),
		withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
		withFixturePosition(&dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 10, Value: 100}),
		withFixtureProjection(&dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 7, Value: 80}),
		withFixtureUser(&dusers.User{Username: "alice"}),
	)

	_, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 30, Outcome: "YES"})
	if !errors.Is(err, bets.ErrInsufficientShares) {
		t.Fatalf("expected inflated projected value to return ErrInsufficientShares, got %v", err)
	}
	if fixture.repo.created != nil || len(fixture.users.calls) != 0 {
		t.Fatalf("inflated projection must not mutate ledger: repo=%+v users=%+v", fixture.repo.created, fixture.users.calls)
	}
}

func TestServiceSell_AttachmentSequenceRejectsOvercashoutBeforeTinyTail(t *testing.T) {
	now := serviceTestTime()
	var history []boundary.Bet
	nextNow := now
	clock := fixedClock{nowFunc: func() time.Time {
		nextNow = nextNow.Add(time.Second)
		return nextNow
	}}
	market := &dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}
	snapshot := positionsmath.MarketSnapshot{ID: market.ID, CreatedAt: now}

	projectPosition := func(extra *boundary.Bet) (*dmarkets.UserPosition, error) {
		projectedHistory := append([]boundary.Bet(nil), history...)
		if extra != nil {
			projectedHistory = append(projectedHistory, *extra)
		}
		position, err := positionsmath.CalculateMarketPositionForUser_WPAM_DBPM(snapshot, projectedHistory, "alice")
		if err != nil {
			return nil, err
		}
		return &dmarkets.UserPosition{
			Username:         "alice",
			MarketID:         market.ID,
			YesSharesOwned:   position.YesSharesOwned,
			NoSharesOwned:    position.NoSharesOwned,
			Value:            position.Value,
			TotalSpent:       position.TotalSpent,
			TotalSpentInPlay: position.TotalSpentInPlay,
			IsResolved:       position.IsResolved,
			ResolutionResult: position.ResolutionResult,
		}, nil
	}

	repo := newFakeRepo(
		withFakeRepoCreate(func(_ context.Context, bet *boundary.Bet) error {
			copied := *bet
			history = append(history, copied)
			return nil
		}),
		withFakeRepoHasBet(func(_ context.Context, marketID uint, username string) (bool, error) {
			for _, bet := range history {
				if bet.MarketID == marketID && bet.Username == username {
					return true, nil
				}
			}
			return false, nil
		}),
	)
	markets := newFakeMarkets(
		withFakeMarket(func(context.Context, int64) (*dmarkets.Market, error) { return market, nil }),
		withFakePosition(func(context.Context, int64, string) (*dmarkets.UserPosition, error) {
			return projectPosition(nil)
		}),
		withFakePositionProjection(func(_ context.Context, _ int64, _ string, bet boundary.Bet) (*dmarkets.UserPosition, error) {
			return projectPosition(&bet)
		}),
	)
	users := newFakeUsers(
		withFakeUserLookup(func(context.Context, string) (*dusers.User, error) {
			return &dusers.User{Username: "alice", AccountBalance: 10000}, nil
		}),
		withFakeApplyTransaction(func(context.Context, string, int64, string) error { return nil }),
	)
	svc := bets.NewService(
		repo,
		markets,
		users,
		bets.Config{MaxDustPerSale: 1, MaximumDebtAllowed: 10000},
		clock,
		bets.WithPlaceUnitOfWork(fakePlaceUnit{repo: repo, users: users}),
		bets.WithSellUnitOfWork(fakeSellUnit{repo: repo, markets: markets, users: users}),
	)

	steps := []struct {
		seq     int
		kind    string
		outcome string
		amount  int64
	}{
		{seq: 1, kind: "buy", outcome: "NO", amount: 50},
		{seq: 2, kind: "buy", outcome: "YES", amount: 100},
		{seq: 3, kind: "buy", outcome: "YES", amount: 50},
		{seq: 4, kind: "sell", outcome: "YES", amount: 200},
		{seq: 5, kind: "buy", outcome: "YES", amount: 100},
		{seq: 6, kind: "buy", outcome: "NO", amount: 10},
		{seq: 7, kind: "buy", outcome: "NO", amount: 40},
		{seq: 8, kind: "buy", outcome: "YES", amount: 100},
		{seq: 9, kind: "buy", outcome: "NO", amount: 50},
		{seq: 10, kind: "sell", outcome: "YES", amount: 400},
		{seq: 11, kind: "buy", outcome: "NO", amount: 200},
		{seq: 12, kind: "buy", outcome: "YES", amount: 100},
		{seq: 13, kind: "sell", outcome: "NO", amount: 600},
		{seq: 14, kind: "buy", outcome: "NO", amount: 10},
		{seq: 15, kind: "buy", outcome: "NO", amount: 10},
		{seq: 16, kind: "sell", outcome: "NO", amount: 520},
		{seq: 17, kind: "buy", outcome: "NO", amount: 10},
		{seq: 18, kind: "buy", outcome: "NO", amount: 10},
	}
	for _, step := range steps {
		switch step.kind {
		case "buy":
			if _, err := svc.Place(context.Background(), bets.PlaceRequest{Username: "alice", MarketID: 1, Amount: step.amount, Outcome: step.outcome}); err != nil {
				t.Fatalf("setup buy seq %d failed: %v", step.seq, err)
			}
		case "sell":
			if _, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: step.amount, Outcome: step.outcome}); err != nil {
				t.Fatalf("setup sell seq %d failed: %v", step.seq, err)
			}
		}
	}

	beforeRows := len(history)
	_, quoteErr := svc.QuoteSell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 507, Outcome: "NO"})
	if !errors.Is(quoteErr, bets.ErrInsufficientShares) {
		t.Fatalf("expected quote for attachment seq 19 over-cashout to return ErrInsufficientShares, got %v", quoteErr)
	}

	_, sellErr := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 507, Outcome: "NO"})
	if !errors.Is(sellErr, bets.ErrInsufficientShares) {
		t.Fatalf("expected sell for attachment seq 19 over-cashout to return ErrInsufficientShares, got %v", sellErr)
	}
	if len(history) != beforeRows {
		t.Fatalf("rejected over-cashout must not append a sale row: before=%d after=%d", beforeRows, len(history))
	}
}

func TestServiceSell_DustScenarioMatrix(t *testing.T) {
	now := serviceTestTime()
	tests := []struct {
		name            string
		maxDust         int64
		sharesOwned     int64
		positionValue   int64
		requestedAmount int64
		wantShares      int64
		wantSaleValue   int64
		wantDust        int64
	}{
		{
			name:            "exact share multiple has zero dust",
			maxDust:         2,
			sharesOwned:     10,
			positionValue:   100,
			requestedAmount: 30,
			wantShares:      3,
			wantSaleValue:   30,
			wantDust:        0,
		},
		{
			name:            "one dust below cap is recorded and withheld",
			maxDust:         2,
			sharesOwned:     10,
			positionValue:   100,
			requestedAmount: 31,
			wantShares:      3,
			wantSaleValue:   30,
			wantDust:        1,
		},
		{
			name:            "dust at cap is allowed",
			maxDust:         2,
			sharesOwned:     10,
			positionValue:   100,
			requestedAmount: 32,
			wantShares:      3,
			wantSaleValue:   30,
			wantDust:        2,
		},
		{
			name:            "dust above cap is rounded down to cap",
			maxDust:         2,
			sharesOwned:     10,
			positionValue:   100,
			requestedAmount: 33,
			wantShares:      3,
			wantSaleValue:   30,
			wantDust:        2,
		},
		{
			name:            "zero cap rounds dust down to zero",
			maxDust:         0,
			sharesOwned:     10,
			positionValue:   100,
			requestedAmount: 35,
			wantShares:      3,
			wantSaleValue:   30,
			wantDust:        0,
		},
		{
			name:            "full-position exact sale has zero dust",
			maxDust:         2,
			sharesOwned:     10,
			positionValue:   100,
			requestedAmount: 100,
			wantShares:      10,
			wantSaleValue:   100,
			wantDust:        0,
		},
		{
			name:            "full-position sale plus cap dust is allowed",
			maxDust:         2,
			sharesOwned:     10,
			positionValue:   100,
			requestedAmount: 102,
			wantShares:      10,
			wantSaleValue:   100,
			wantDust:        2,
		},
		{
			name:            "full-position sale above cap is rounded down to cap",
			maxDust:         2,
			sharesOwned:     10,
			positionValue:   100,
			requestedAmount: 103,
			wantShares:      10,
			wantSaleValue:   100,
			wantDust:        2,
		},
		{
			name:            "larger share value creates larger possible dust interval",
			maxDust:         4,
			sharesOwned:     5,
			positionValue:   100,
			requestedAmount: 44,
			wantShares:      2,
			wantSaleValue:   40,
			wantDust:        4,
		},
		{
			name:            "larger share value rounds dust beyond cap down",
			maxDust:         4,
			sharesOwned:     5,
			positionValue:   100,
			requestedAmount: 45,
			wantShares:      2,
			wantSaleValue:   40,
			wantDust:        4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture, svc := newServiceFixture(
				now,
				withFixtureMaxDust(tc.maxDust),
				withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
				withFixturePosition(&dmarkets.UserPosition{
					Username:       "alice",
					MarketID:       1,
					YesSharesOwned: tc.sharesOwned,
					Value:          tc.positionValue,
				}),
				withFixtureUser(&dusers.User{Username: "alice"}),
			)

			result, err := svc.Sell(context.Background(), bets.SellRequest{
				Username: "alice",
				MarketID: 1,
				Amount:   tc.requestedAmount,
				Outcome:  "YES",
			})

			if err != nil {
				t.Fatalf("Sell returned error: %v", err)
			}
			wantNetProceeds := tc.wantSaleValue - tc.wantDust
			if result.SharesSold != tc.wantShares || result.SaleValue != tc.wantSaleValue || result.Dust != tc.wantDust || result.NetProceeds != wantNetProceeds {
				t.Fatalf("unexpected result: got %+v, want shares=%d saleValue=%d dust=%d netProceeds=%d", result, tc.wantShares, tc.wantSaleValue, tc.wantDust, wantNetProceeds)
			}
			if fixture.repo.created == nil {
				t.Fatal("expected stored sale bet")
			}
			if fixture.repo.created.Amount != -tc.wantShares {
				t.Fatalf("unexpected stored sale bet: %+v", fixture.repo.created)
			}
			if len(fixture.users.calls) != 1 || fixture.users.calls[0].transaction != dusers.TransactionSale || fixture.users.calls[0].amount != wantNetProceeds {
				t.Fatalf("unexpected user ledger calls: %+v", fixture.users.calls)
			}
		})
	}
}

func TestServiceSell_MarketHistorySaleOrderDustScenarios(t *testing.T) {
	now := serviceTestTime()
	priorHistory := []boundary.Bet{
		{Username: "alice", MarketID: 1, Amount: 70, Outcome: "YES", PlacedAt: now.Add(-4 * time.Hour)},
		{Username: "bob", MarketID: 1, Amount: 40, Outcome: "NO", PlacedAt: now.Add(-3 * time.Hour)},
		{Username: "alice", MarketID: 1, Amount: 30, Outcome: "YES", PlacedAt: now.Add(-2 * time.Hour)},
		{Username: "bob", MarketID: 1, Amount: -1, Outcome: "NO", PlacedAt: now.Add(-1 * time.Hour)},
	}

	tests := []struct {
		name                string
		maxDust             int64
		requestedAmount     int64
		wantExecutableOrder int64
		wantShares          int64
		wantSaleValue       int64
		wantDust            int64
	}{
		{
			name:                "prior buy sell history then zero dust sale order",
			maxDust:             1,
			requestedAmount:     30,
			wantExecutableOrder: 30,
			wantShares:          3,
			wantSaleValue:       30,
			wantDust:            0,
		},
		{
			name:                "prior buy sell history then one dust sale order executes",
			maxDust:             1,
			requestedAmount:     31,
			wantExecutableOrder: 31,
			wantShares:          3,
			wantSaleValue:       30,
			wantDust:            1,
		},
		{
			name:                "prior buy sell history then over remainder rounds down to one dust",
			maxDust:             1,
			requestedAmount:     35,
			wantExecutableOrder: 31,
			wantShares:          3,
			wantSaleValue:       30,
			wantDust:            1,
		},
		{
			name:                "prior buy sell history then over remainder rounds down to zero dust when cap is zero",
			maxDust:             0,
			requestedAmount:     35,
			wantExecutableOrder: 30,
			wantShares:          3,
			wantSaleValue:       30,
			wantDust:            0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			position := positionFromHistoryForAlice(priorHistory)
			fixture, svc := newServiceFixture(
				now,
				withFixtureMaxDust(tc.maxDust),
				withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
				withFixturePosition(position),
				withFixtureUser(&dusers.User{Username: "alice"}),
			)

			quote, err := svc.QuoteSell(context.Background(), bets.SellRequest{
				Username: "alice",
				MarketID: 1,
				Amount:   tc.requestedAmount,
				Outcome:  "YES",
			})
			if err != nil {
				t.Fatalf("QuoteSell returned error: %v", err)
			}
			wantNetProceeds := tc.wantSaleValue - tc.wantDust
			if !quote.Allowed || quote.RequestedCredits != tc.wantExecutableOrder || quote.SharesSold != tc.wantShares || quote.SaleValue != tc.wantSaleValue || quote.Dust != tc.wantDust || quote.NetProceeds != wantNetProceeds {
				t.Fatalf("unexpected quote: got %+v, want order=%d shares=%d saleValue=%d dust=%d netProceeds=%d", quote, tc.wantExecutableOrder, tc.wantShares, tc.wantSaleValue, tc.wantDust, wantNetProceeds)
			}
			if fixture.repo.created != nil || len(fixture.users.calls) != 0 {
				t.Fatalf("quote should not mutate state: repo=%+v users=%+v", fixture.repo.created, fixture.users.calls)
			}

			result, err := svc.Sell(context.Background(), bets.SellRequest{
				Username: "alice",
				MarketID: 1,
				Amount:   tc.requestedAmount,
				Outcome:  "YES",
			})
			if err != nil {
				t.Fatalf("Sell returned error: %v", err)
			}
			if result.SharesSold != tc.wantShares || result.SaleValue != tc.wantSaleValue || result.Dust != tc.wantDust || result.NetProceeds != wantNetProceeds {
				t.Fatalf("unexpected sale result: got %+v, want shares=%d saleValue=%d dust=%d netProceeds=%d", result, tc.wantShares, tc.wantSaleValue, tc.wantDust, wantNetProceeds)
			}
			if fixture.repo.created == nil || fixture.repo.created.Amount != -tc.wantShares || fixture.repo.created.Outcome != "YES" {
				t.Fatalf("unexpected stored sale bet: %+v", fixture.repo.created)
			}
			if len(fixture.users.calls) != 1 || fixture.users.calls[0].transaction != dusers.TransactionSale || fixture.users.calls[0].amount != wantNetProceeds {
				t.Fatalf("unexpected user ledger calls: %+v", fixture.users.calls)
			}
		})
	}
}

func TestServiceSell_ActualMixedMarketHistorySaleOrderDustScenarios(t *testing.T) {
	now := serviceTestTime()
	history := []boundary.Bet{
		{Username: "alice", MarketID: 1, Amount: 10, Outcome: "YES", PlacedAt: now.Add(-5 * time.Hour), CreatedAt: now.Add(-5 * time.Hour)},
		{Username: "bob", MarketID: 1, Amount: 10000, Outcome: "NO", PlacedAt: now.Add(-4 * time.Hour), CreatedAt: now.Add(-4 * time.Hour)},
		{Username: "carol", MarketID: 1, Amount: 10000, Outcome: "NO", PlacedAt: now.Add(-3 * time.Hour), CreatedAt: now.Add(-3 * time.Hour)},
		{Username: "dave", MarketID: 1, Amount: 50000, Outcome: "YES", PlacedAt: now.Add(-2 * time.Hour), CreatedAt: now.Add(-2 * time.Hour)},
		{Username: "erin", MarketID: 1, Amount: 50000, Outcome: "YES", PlacedAt: now.Add(-1 * time.Hour), CreatedAt: now.Add(-1 * time.Hour)},
	}
	snapshot := positionsmath.MarketSnapshot{ID: 1, CreatedAt: now.Add(-6 * time.Hour)}
	position, err := positionsmath.CalculateMarketPositionForUser_WPAM_DBPM(snapshot, history, "alice")
	if err != nil {
		t.Fatalf("calculate actual position: %v", err)
	}
	if position.YesSharesOwned != 20 || position.Value != 8350 {
		t.Fatalf("expected mixed market history to produce Alice YES position worth 8350 over 20 shares, got %+v", position)
	}
	valuePerShare := position.Value / position.YesSharesOwned
	if valuePerShare != 417 {
		t.Fatalf("expected mixed market history valuePerShare=417, got %d from position %+v", valuePerShare, position)
	}

	tests := []struct {
		name                string
		maxDust             int64
		requestedAmount     int64
		wantExecutableOrder int64
		wantShares          int64
		wantSaleValue       int64
		wantDust            int64
	}{
		{
			name:                "actual mixed history then zero dust sale order",
			maxDust:             1,
			requestedAmount:     1251,
			wantExecutableOrder: 1251,
			wantShares:          3,
			wantSaleValue:       1251,
			wantDust:            0,
		},
		{
			name:                "actual mixed history then one dust sale order executes",
			maxDust:             1,
			requestedAmount:     1252,
			wantExecutableOrder: 1252,
			wantShares:          3,
			wantSaleValue:       1251,
			wantDust:            1,
		},
		{
			name:                "actual mixed history then over remainder rounds down to one dust",
			maxDust:             1,
			requestedAmount:     1255,
			wantExecutableOrder: 1252,
			wantShares:          3,
			wantSaleValue:       1251,
			wantDust:            1,
		},
		{
			name:                "actual mixed history then over remainder rounds down to zero dust when cap is zero",
			maxDust:             0,
			requestedAmount:     1255,
			wantExecutableOrder: 1251,
			wantShares:          3,
			wantSaleValue:       1251,
			wantDust:            0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture, svc := newServiceFixture(
				now,
				withFixtureMaxDust(tc.maxDust),
				withFixtureMarket(&dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}),
				withFixturePosition(&dmarkets.UserPosition{
					Username:         "alice",
					MarketID:         1,
					YesSharesOwned:   position.YesSharesOwned,
					NoSharesOwned:    position.NoSharesOwned,
					Value:            position.Value,
					TotalSpent:       position.TotalSpent,
					TotalSpentInPlay: position.TotalSpentInPlay,
				}),
				withFixtureUser(&dusers.User{Username: "alice"}),
			)

			quote, err := svc.QuoteSell(context.Background(), bets.SellRequest{
				Username: "alice",
				MarketID: 1,
				Amount:   tc.requestedAmount,
				Outcome:  "YES",
			})
			if err != nil {
				t.Fatalf("QuoteSell returned error: %v", err)
			}
			wantNetProceeds := tc.wantSaleValue - tc.wantDust
			if !quote.Allowed || quote.RequestedCredits != tc.wantExecutableOrder || quote.SharesSold != tc.wantShares || quote.SaleValue != tc.wantSaleValue || quote.Dust != tc.wantDust || quote.NetProceeds != wantNetProceeds {
				t.Fatalf("unexpected quote: got %+v, want order=%d shares=%d saleValue=%d dust=%d netProceeds=%d", quote, tc.wantExecutableOrder, tc.wantShares, tc.wantSaleValue, tc.wantDust, wantNetProceeds)
			}
			if fixture.repo.created != nil || len(fixture.users.calls) != 0 {
				t.Fatalf("quote should not mutate state: repo=%+v users=%+v", fixture.repo.created, fixture.users.calls)
			}

			result, err := svc.Sell(context.Background(), bets.SellRequest{
				Username: "alice",
				MarketID: 1,
				Amount:   tc.requestedAmount,
				Outcome:  "YES",
			})
			if err != nil {
				t.Fatalf("Sell returned error: %v", err)
			}
			if result.SharesSold != tc.wantShares || result.SaleValue != tc.wantSaleValue || result.Dust != tc.wantDust || result.NetProceeds != wantNetProceeds {
				t.Fatalf("unexpected sale result: got %+v, want shares=%d saleValue=%d dust=%d netProceeds=%d", result, tc.wantShares, tc.wantSaleValue, tc.wantDust, wantNetProceeds)
			}
			if fixture.repo.created == nil || fixture.repo.created.Amount != -tc.wantShares || fixture.repo.created.Outcome != "YES" {
				t.Fatalf("unexpected stored sale bet: %+v", fixture.repo.created)
			}
			if len(fixture.users.calls) != 1 || fixture.users.calls[0].transaction != dusers.TransactionSale || fixture.users.calls[0].amount != wantNetProceeds {
				t.Fatalf("unexpected user ledger calls: %+v", fixture.users.calls)
			}
		})
	}
}

func positionFromHistoryForAlice(history []boundary.Bet) *dmarkets.UserPosition {
	var yesShares int64
	var value int64
	for _, bet := range history {
		if bet.Username != "alice" || bet.Outcome != "YES" {
			continue
		}
		yesShares += bet.Amount / 10
		value += bet.Amount
	}
	return &dmarkets.UserPosition{
		Username:       "alice",
		MarketID:       1,
		YesSharesOwned: yesShares,
		Value:          value,
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

func containsInt64(values []int64, want int64) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
