package bets

import (
	"context"
	"errors"
	"testing"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/models"
	"socialpredict/setup"
)

type stubMarketService struct {
	market *dmarkets.Market
	err    error
}

func (s stubMarketService) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.market, nil
}

func (s stubMarketService) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	return nil, errors.New("unexpected call")
}

type gateClock struct{ now time.Time }

func (c gateClock) Now() time.Time { return c.now }

func TestMarketGate_Open(t *testing.T) {
	now := time.Now()
	openGate := marketGate{markets: stubMarketService{market: &dmarkets.Market{Status: "active", ResolutionDateTime: now.Add(time.Hour)}}, clock: gateClock{now: now}}
	if _, err := openGate.Open(context.Background(), 1); err != nil {
		t.Fatalf("expected open market, got %v", err)
	}

	resolvedGate := marketGate{markets: stubMarketService{market: &dmarkets.Market{Status: "resolved", ResolutionDateTime: now.Add(-time.Hour)}}, clock: gateClock{now: now}}
	if _, err := resolvedGate.Open(context.Background(), 1); !errors.Is(err, ErrMarketClosed) {
		t.Fatalf("expected ErrMarketClosed, got %v", err)
	}

	failingGate := marketGate{markets: stubMarketService{err: errors.New("boom")}, clock: gateClock{now: now}}
	if _, err := failingGate.Open(context.Background(), 1); err == nil {
		t.Fatalf("expected error from market service")
	}
}

func TestFeeCalculator_Calculate(t *testing.T) {
	econ := &setup.EconomicConfig{}
	econ.Economics.Betting.BetFees.InitialBetFee = 5
	econ.Economics.Betting.BetFees.BuySharesFee = 2

	calc := feeCalculator{econ: econ}

	fees := calc.Calculate(false, 10)
	if fees.initialFee != 5 || fees.transactionFee != 2 || fees.totalCost != 17 {
		t.Fatalf("unexpected fees with no prior bet: %+v", fees)
	}

	fees = calc.Calculate(true, 10)
	if fees.initialFee != 0 || fees.transactionFee != 2 || fees.totalCost != 12 {
		t.Fatalf("unexpected fees with prior bet: %+v", fees)
	}
}

func TestBalanceGuard_EnsureSufficient(t *testing.T) {
	guard := balanceGuard{maxDebtAllowed: 50}

	if err := guard.EnsureSufficient(0, 40); err != nil {
		t.Fatalf("expected balance to pass: %v", err)
	}

	if err := guard.EnsureSufficient(-10, 100); !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
}

type ledgerRepo struct {
	bet       *models.Bet
	createErr error
}

func (l *ledgerRepo) Create(ctx context.Context, bet *models.Bet) error {
	if l.createErr != nil {
		return l.createErr
	}
	copyBet := *bet
	l.bet = &copyBet
	return nil
}

func (l *ledgerRepo) UserHasBet(ctx context.Context, marketID uint, username string) (bool, error) {
	return false, errors.New("unexpected call")
}

type ledgerCall struct {
	username    string
	amount      int64
	transaction string
}

type ledgerUsers struct {
	calls    []ledgerCall
	applyErr error
}

func (u *ledgerUsers) GetUser(ctx context.Context, username string) (*dusers.User, error) {
	return nil, errors.New("unexpected call")
}

func (u *ledgerUsers) ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error {
	if u.applyErr != nil {
		return u.applyErr
	}
	u.calls = append(u.calls, ledgerCall{username: username, amount: amount, transaction: transactionType})
	return nil
}

func TestBetLedger_ChargeAndRecord(t *testing.T) {
	users := &ledgerUsers{}
	repo := &ledgerRepo{}
	ledger := betLedger{repo: repo, users: users}
	bet := &models.Bet{Username: "bob"}

	if err := ledger.ChargeAndRecord(context.Background(), bet, 25); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(users.calls) != 1 || users.calls[0].transaction != dusers.TransactionBuy || users.calls[0].amount != 25 {
		t.Fatalf("unexpected user calls: %+v", users.calls)
	}
	if repo.bet == nil {
		t.Fatalf("expected bet persisted")
	}
}

func TestBetLedger_ChargeAndRecord_RollsBackOnRepoError(t *testing.T) {
	users := &ledgerUsers{}
	repo := &ledgerRepo{createErr: errors.New("db down")}
	ledger := betLedger{repo: repo, users: users}

	err := ledger.ChargeAndRecord(context.Background(), &models.Bet{Username: "alice"}, 10)
	if err == nil {
		t.Fatalf("expected error")
	}
	if len(users.calls) != 2 || users.calls[1].transaction != dusers.TransactionRefund {
		t.Fatalf("expected refund on failure, calls: %+v", users.calls)
	}
}

func TestBetLedger_CreditSale(t *testing.T) {
	users := &ledgerUsers{}
	repo := &ledgerRepo{}
	ledger := betLedger{repo: repo, users: users}

	if err := ledger.CreditSale(context.Background(), &models.Bet{Username: "alice"}, 15); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(users.calls) != 1 || users.calls[0].transaction != dusers.TransactionSale || users.calls[0].amount != 15 {
		t.Fatalf("unexpected user calls: %+v", users.calls)
	}
}

func TestSaleCalculator_Calculate(t *testing.T) {
	calc := saleCalculator{maxDustPerSale: 3}
	pos := &dmarkets.UserPosition{Value: 100}

	result, err := calc.Calculate(pos, 10, 23)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if result.sharesToSell != 2 || result.saleValue != 20 || result.dust != 3 {
		t.Fatalf("unexpected sale result: %+v", result)
	}

	if _, err := calc.Calculate(pos, 10, 5); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}

	_, err = calc.Calculate(pos, 10, 35)
	var dustErr ErrDustCapExceeded
	if !errors.As(err, &dustErr) {
		t.Fatalf("expected dust cap error, got %v", err)
	}
}
