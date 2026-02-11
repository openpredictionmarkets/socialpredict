package bets_test

import (
	"context"
	"errors"
	"testing"
	"time"

	bets "socialpredict/internal/domain/bets"
	dmarkets "socialpredict/internal/domain/markets"
	dwallet "socialpredict/internal/domain/wallet"
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

type walletCall struct {
	kind        string
	username    string
	amount      int64
	maxDebt     int64
	transaction string
}

type fakeWallet struct {
	validateErr error
	debitErr    error
	creditErr   error
	calls       []walletCall
}

func (f *fakeWallet) ValidateBalance(ctx context.Context, username string, amount int64, maxDebt int64) error {
	f.calls = append(f.calls, walletCall{
		kind:     "validate",
		username: username,
		amount:   amount,
		maxDebt:  maxDebt,
	})
	return f.validateErr
}

func (f *fakeWallet) Debit(ctx context.Context, username string, amount int64, maxDebt int64, txType string) error {
	f.calls = append(f.calls, walletCall{
		kind:        "debit",
		username:    username,
		amount:      amount,
		maxDebt:     maxDebt,
		transaction: txType,
	})
	return f.debitErr
}

func (f *fakeWallet) Credit(ctx context.Context, username string, amount int64, txType string) error {
	f.calls = append(f.calls, walletCall{
		kind:        "credit",
		username:    username,
		amount:      amount,
		transaction: txType,
	})
	return f.creditErr
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func TestServicePlace_Succeeds(t *testing.T) {
	econ := modelstesting.GenerateEconomicConfig()
	now := time.Now()
	expectedMaxDebt := int64(econ.Economics.User.MaximumDebtAllowed)

	repo := &fakeRepo{}
	markets := &fakeMarkets{market: &dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}}
	wallet := &fakeWallet{}

	svc := bets.NewServiceWithWallet(repo, markets, wallet, econ, fixedClock{now: now})

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

	totalCost := int64(100 + econ.Economics.Betting.BetFees.InitialBetFee + econ.Economics.Betting.BetFees.BuySharesFee)
	if len(wallet.calls) != 2 {
		t.Fatalf("expected validate+debit wallet calls, got %+v", wallet.calls)
	}
	if wallet.calls[0].kind != "validate" || wallet.calls[0].amount != totalCost || wallet.calls[0].maxDebt != expectedMaxDebt {
		t.Fatalf("unexpected validate wallet call: %+v", wallet.calls[0])
	}
	if wallet.calls[1].kind != "debit" || wallet.calls[1].amount != totalCost || wallet.calls[1].maxDebt != expectedMaxDebt || wallet.calls[1].transaction != dwallet.TxBuy {
		t.Fatalf("unexpected debit wallet call: %+v", wallet.calls[1])
	}
}

func TestServicePlace_InsufficientBalance(t *testing.T) {
	econ := modelstesting.GenerateEconomicConfig()
	now := time.Now()

	repo := &fakeRepo{}
	markets := &fakeMarkets{market: &dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}}
	wallet := &fakeWallet{validateErr: dwallet.ErrInsufficientBalance}

	svc := bets.NewServiceWithWallet(repo, markets, wallet, econ, fixedClock{now: now})

	_, err := svc.Place(context.Background(), bets.PlaceRequest{Username: "alice", MarketID: 1, Amount: 9999, Outcome: "YES"})
	if !errors.Is(err, bets.ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
	if repo.created != nil {
		t.Fatalf("expected no persisted bet on insufficient balance")
	}
}

func TestServicePlace_InvalidOutcome(t *testing.T) {
	econ := modelstesting.GenerateEconomicConfig()
	now := time.Now()

	repo := &fakeRepo{}
	markets := &fakeMarkets{market: &dmarkets.Market{ID: 1, Status: "active", ResolutionDateTime: now.Add(24 * time.Hour)}}
	wallet := &fakeWallet{}

	svc := bets.NewServiceWithWallet(repo, markets, wallet, econ, fixedClock{now: now})

	_, err := svc.Place(context.Background(), bets.PlaceRequest{Username: "alice", MarketID: 1, Amount: 10, Outcome: "MAYBE"})
	if !errors.Is(err, bets.ErrInvalidOutcome) {
		t.Fatalf("expected ErrInvalidOutcome, got %v", err)
	}
	if len(wallet.calls) != 0 {
		t.Fatalf("expected no wallet calls on invalid input, got %+v", wallet.calls)
	}
}

func TestServicePlace_MarketClosed(t *testing.T) {
	econ := modelstesting.GenerateEconomicConfig()
	now := time.Now()

	repo := &fakeRepo{}
	markets := &fakeMarkets{market: &dmarkets.Market{ID: 1, Status: "resolved", ResolutionDateTime: now.Add(-time.Hour)}}
	wallet := &fakeWallet{}

	svc := bets.NewServiceWithWallet(repo, markets, wallet, econ, fixedClock{now: now})

	_, err := svc.Place(context.Background(), bets.PlaceRequest{Username: "alice", MarketID: 1, Amount: 10, Outcome: "YES"})
	if !errors.Is(err, bets.ErrMarketClosed) {
		t.Fatalf("expected ErrMarketClosed, got %v", err)
	}
	if len(wallet.calls) != 0 {
		t.Fatalf("expected no wallet calls when market is closed, got %+v", wallet.calls)
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
	wallet := &fakeWallet{}

	svc := bets.NewServiceWithWallet(repo, markets, wallet, econ, fixedClock{now: now})

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
	if len(wallet.calls) != 1 || wallet.calls[0].kind != "credit" || wallet.calls[0].transaction != dwallet.TxSale || wallet.calls[0].amount != 20 {
		t.Fatalf("unexpected wallet calls: %+v", wallet.calls)
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
	wallet := &fakeWallet{}

	svc := bets.NewServiceWithWallet(repo, markets, wallet, econ, fixedClock{now: now})

	_, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 10, Outcome: "YES"})
	if !errors.Is(err, bets.ErrNoPosition) {
		t.Fatalf("expected ErrNoPosition, got %v", err)
	}
	if len(wallet.calls) != 0 {
		t.Fatalf("expected no wallet calls when no position exists, got %+v", wallet.calls)
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
	wallet := &fakeWallet{}

	svc := bets.NewServiceWithWallet(repo, markets, wallet, econ, fixedClock{now: now})

	_, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 33, Outcome: "YES"})
	if _, ok := err.(bets.ErrDustCapExceeded); !ok {
		t.Fatalf("expected ErrDustCapExceeded, got %v", err)
	}
	if len(wallet.calls) != 0 {
		t.Fatalf("expected no wallet calls on dust cap failure, got %+v", wallet.calls)
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
	wallet := &fakeWallet{}

	svc := bets.NewServiceWithWallet(repo, markets, wallet, econ, fixedClock{now: now})

	_, err := svc.Sell(context.Background(), bets.SellRequest{
		Username: "alice",
		MarketID: 1,
		Amount:   5, // less than value per share (10)
		Outcome:  "YES",
	})
	if !errors.Is(err, bets.ErrInvalidAmount) {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}
	if len(wallet.calls) != 0 {
		t.Fatalf("expected no wallet calls for invalid sell request, got %+v", wallet.calls)
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
	wallet := &fakeWallet{creditErr: errors.New("wallet unavailable")}

	svc := bets.NewServiceWithWallet(repo, markets, wallet, econ, fixedClock{now: now})

	_, err := svc.Sell(context.Background(), bets.SellRequest{Username: "alice", MarketID: 1, Amount: 25, Outcome: "YES"})
	if err == nil {
		t.Fatalf("expected error from wallet credit failure")
	}
	if repo.created != nil {
		t.Fatalf("expected no bet persisted when wallet credit fails")
	}
}
