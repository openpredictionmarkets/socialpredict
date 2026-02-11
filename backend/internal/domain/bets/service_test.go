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
)

type fakeRepo struct {
	created   *models.Bet
	createErr error
	hasBet    bool
	hasErr    error
}

func (f *fakeRepo) Create(ctx context.Context, bet *models.Bet) error {
	if f.createErr != nil {
		return f.createErr
	}
	copied := *bet
	f.created = &copied
	return nil
}

func (f *fakeRepo) UserHasBet(ctx context.Context, marketID uint, username string) (bool, error) {
	if f.hasErr != nil {
		return false, f.hasErr
	}
	return f.hasBet, nil
}

type fakeMarkets struct {
	market     *dmarkets.Market
	marketErr  error
	userPos    *dmarkets.UserPosition
	userPosErr error
}

func (f *fakeMarkets) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	if f.marketErr != nil {
		return nil, f.marketErr
	}
	return f.market, nil
}

func (f *fakeMarkets) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	if f.userPosErr != nil {
		return nil, f.userPosErr
	}
	return f.userPos, nil
}

type applyCall struct {
	username    string
	amount      int64
	transaction string
}

type fakeUsers struct {
	user     *dusers.User
	getErr   error
	applyErr error
	calls    []applyCall
}

func (f *fakeUsers) GetUser(ctx context.Context, username string) (*dusers.User, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.user, nil
}

func (f *fakeUsers) ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error {
	if f.applyErr != nil {
		return f.applyErr
	}
	f.calls = append(f.calls, applyCall{username: username, amount: amount, transaction: transactionType})
	return nil
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func TestServicePlace_Succeeds(t *testing.T) {
	econ := modelstesting.GenerateEconomicConfig()
	now := time.Now()

	repo := &fakeRepo{}
	markets := &fakeMarkets{market: &dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}}
	users := &fakeUsers{user: &dusers.User{Username: "alice", AccountBalance: 500}}

	svc := bets.NewService(repo, markets, users, econ, fixedClock{now: now})

	placed, err := svc.Place(context.Background(), bets.PlaceRequest{Username: "alice", MarketID: 1, Amount: 100, Outcome: "yes"})
	if err != nil {
		t.Fatalf("Place returned error: %v", err)
	}

	if placed.Username != "alice" || placed.Amount != 100 || placed.MarketID != 1 {
		t.Fatalf("unexpected placed bet: %+v", placed)
	}

	if repo.created == nil {
		t.Fatalf("expected repository Create to be called")
	}
	if repo.created.Outcome != "YES" {
		t.Fatalf("expected outcome YES, got %s", repo.created.Outcome)
	}

	if len(users.calls) != 1 {
		t.Fatalf("expected one ApplyTransaction call, got %d", len(users.calls))
	}
	totalCost := int64(100 + econ.Economics.Betting.BetFees.InitialBetFee + econ.Economics.Betting.BetFees.BuySharesFee)
	if users.calls[0].amount != totalCost {
		t.Fatalf("unexpected transaction amount: %d", users.calls[0].amount)
	}
}

func TestServicePlace_InsufficientBalance(t *testing.T) {
	econ := modelstesting.GenerateEconomicConfig()
	now := time.Now()

	repo := &fakeRepo{}
	markets := &fakeMarkets{market: &dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}}
	users := &fakeUsers{user: &dusers.User{Username: "alice", AccountBalance: 0}}

	svc := bets.NewService(repo, markets, users, econ, fixedClock{now: now})

	_, err := svc.Place(context.Background(), bets.PlaceRequest{Username: "alice", MarketID: 1, Amount: 9999, Outcome: "YES"})
	if !errors.Is(err, bets.ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
}

func TestServicePlace_InvalidOutcome(t *testing.T) {
	econ := modelstesting.GenerateEconomicConfig()
	now := time.Now()

	repo := &fakeRepo{}
	markets := &fakeMarkets{market: &dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}}
	users := &fakeUsers{user: &dusers.User{Username: "alice", AccountBalance: 100}}

	svc := bets.NewService(repo, markets, users, econ, fixedClock{now: now})

	_, err := svc.Place(context.Background(), bets.PlaceRequest{Username: "alice", MarketID: 1, Amount: 10, Outcome: "MAYBE"})
	if !errors.Is(err, bets.ErrInvalidOutcome) {
		t.Fatalf("expected ErrInvalidOutcome, got %v", err)
	}
}

func TestServicePlace_MarketClosed(t *testing.T) {
	econ := modelstesting.GenerateEconomicConfig()
	now := time.Now()

	repo := &fakeRepo{}
	markets := &fakeMarkets{market: &dmarkets.Market{ID: 1, Status: "resolved", ResolutionDateTime: now.Add(-time.Hour)}}
	users := &fakeUsers{user: &dusers.User{Username: "alice", AccountBalance: 100}}

	svc := bets.NewService(repo, markets, users, econ, fixedClock{now: now})

	_, err := svc.Place(context.Background(), bets.PlaceRequest{Username: "alice", MarketID: 1, Amount: 10, Outcome: "YES"})
	if !errors.Is(err, bets.ErrMarketClosed) {
		t.Fatalf("expected ErrMarketClosed, got %v", err)
	}
}

func TestServiceSell_Succeeds(t *testing.T) {
	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.Betting.MaxDustPerSale = 0
	now := time.Now()

	repo := &fakeRepo{}
	markets := &fakeMarkets{
		market:  &dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)},
		userPos: &dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 10, NoSharesOwned: 0, Value: 100},
	}
	users := &fakeUsers{user: &dusers.User{Username: "alice"}}

	svc := bets.NewService(repo, markets, users, econ, fixedClock{now: now})

	res, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 25, Outcome: "YES"})
	if err != nil {
		t.Fatalf("Sell returned error: %v", err)
	}
	if res.SharesSold != 2 || res.SaleValue != 20 || res.Dust != 5 {
		t.Fatalf("unexpected sell result: %+v", res)
	}
	if repo.created == nil || repo.created.Amount != -2 || repo.created.Outcome != "YES" {
		t.Fatalf("unexpected stored bet: %+v", repo.created)
	}
	if len(users.calls) != 1 || users.calls[0].transaction != dusers.TransactionSale || users.calls[0].amount != 20 {
		t.Fatalf("unexpected user transaction: %+v", users.calls)
	}
}

func TestServiceSell_NoPosition(t *testing.T) {
	econ := modelstesting.GenerateEconomicConfig()
	now := time.Now()

	repo := &fakeRepo{}
	markets := &fakeMarkets{
		market:  &dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)},
		userPos: &dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 0, NoSharesOwned: 0, Value: 0},
	}
	users := &fakeUsers{user: &dusers.User{Username: "alice"}}

	svc := bets.NewService(repo, markets, users, econ, fixedClock{now: now})

	_, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 10, Outcome: "YES"})
	if !errors.Is(err, bets.ErrNoPosition) {
		t.Fatalf("expected ErrNoPosition, got %v", err)
	}
}

func TestServiceSell_DustCapExceeded(t *testing.T) {
	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.Betting.MaxDustPerSale = 2
	now := time.Now()

	repo := &fakeRepo{}
	markets := &fakeMarkets{
		market:  &dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)},
		userPos: &dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 10, Value: 100},
	}
	users := &fakeUsers{user: &dusers.User{Username: "alice"}}

	svc := bets.NewService(repo, markets, users, econ, fixedClock{now: now})

	_, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 33, Outcome: "YES"})
	if _, ok := err.(bets.ErrDustCapExceeded); !ok {
		t.Fatalf("expected ErrDustCapExceeded, got %v", err)
	}
}

func TestServiceSell_RequestTooSmall(t *testing.T) {
	econ := modelstesting.GenerateEconomicConfig()
	now := time.Now()

	repo := &fakeRepo{}
	markets := &fakeMarkets{
		market: &dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)},
		userPos: &dmarkets.UserPosition{
			Username:       "alice",
			MarketID:       1,
			YesSharesOwned: 5,
			Value:          50,
		},
	}
	users := &fakeUsers{user: &dusers.User{Username: "alice"}}

	svc := bets.NewService(repo, markets, users, econ, fixedClock{now: now})

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

func TestServiceSell_WalletCreditFailure(t *testing.T) {
	econ := modelstesting.GenerateEconomicConfig()
	econ.Economics.Betting.MaxDustPerSale = 0
	now := time.Now()

	repo := &fakeRepo{}
	markets := &fakeMarkets{
		market:  &dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)},
		userPos: &dmarkets.UserPosition{Username: "alice", MarketID: 1, YesSharesOwned: 10, NoSharesOwned: 0, Value: 100},
	}
	users := &fakeUsers{user: &dusers.User{Username: "alice"}, applyErr: errors.New("wallet unavailable")}

	svc := bets.NewService(repo, markets, users, econ, fixedClock{now: now})

	_, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 25, Outcome: "YES"})
	if err == nil {
		t.Fatalf("expected error from wallet credit failure")
	}
	if repo.created != nil {
		t.Fatalf("expected no bet persisted when wallet credit fails")
	}
}
