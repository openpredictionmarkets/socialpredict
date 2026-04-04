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

var errUnexpectedHelperCall = errors.New("unexpected call")

type stubMarketService struct {
	getMarketFunc             func(ctx context.Context, id int64) (*dmarkets.Market, error)
	getUserPositionInMarketFn func(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error)
}

func newStubMarketService(opts ...func(*stubMarketService)) stubMarketService {
	stub := stubMarketService{
		getMarketFunc: func(context.Context, int64) (*dmarkets.Market, error) {
			return nil, errUnexpectedHelperCall
		},
		getUserPositionInMarketFn: func(context.Context, int64, string) (*dmarkets.UserPosition, error) {
			return nil, errUnexpectedHelperCall
		},
	}
	for _, opt := range opts {
		opt(&stub)
	}
	return stub
}

func withStubMarket(getMarket func(ctx context.Context, id int64) (*dmarkets.Market, error)) func(*stubMarketService) {
	return func(stub *stubMarketService) {
		stub.getMarketFunc = getMarket
	}
}

func withStubPosition(getPosition func(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error)) func(*stubMarketService) {
	return func(stub *stubMarketService) {
		stub.getUserPositionInMarketFn = getPosition
	}
}

func (s stubMarketService) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	return s.getMarketFunc(ctx, id)
}

func (s stubMarketService) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	return s.getUserPositionInMarketFn(ctx, marketID, username)
}

type gateClock struct {
	nowFunc func() time.Time
}

func newGateClock(now time.Time) gateClock {
	return gateClock{
		nowFunc: func() time.Time { return now },
	}
}

func (c gateClock) Now() time.Time {
	if c.nowFunc == nil {
		return time.Time{}
	}
	return c.nowFunc()
}

func helperTestTime() time.Time {
	return time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
}

func TestMarketGate_Open(t *testing.T) {
	now := helperTestTime()
	tests := []struct {
		name    string
		market  *dmarkets.Market
		err     error
		wantErr error
	}{
		{
			name:   "open market",
			market: &dmarkets.Market{Status: "active", ResolutionDateTime: now.Add(time.Hour)},
		},
		{
			name:    "resolved market",
			market:  &dmarkets.Market{Status: "resolved", ResolutionDateTime: now.Add(-time.Hour)},
			wantErr: ErrMarketClosed,
		},
		{
			name:    "service error",
			err:     errors.New("boom"),
			wantErr: errors.New("boom"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gate := marketGate{
				markets: newStubMarketService(withStubMarket(func(ctx context.Context, id int64) (*dmarkets.Market, error) {
					return tc.market, tc.err
				})),
				clock: newGateClock(now),
			}

			_, err := gate.Open(context.Background(), 1)
			switch {
			case tc.wantErr == nil && err != nil:
				t.Fatalf("expected success, got %v", err)
			case tc.wantErr != nil && err == nil:
				t.Fatalf("expected error %v", tc.wantErr)
			case tc.name == "service error" && err == nil:
				t.Fatalf("expected service error")
			case tc.name == "service error" && err.Error() != tc.err.Error():
				t.Fatalf("expected propagated error %q, got %v", tc.err.Error(), err)
			case tc.wantErr != nil && tc.name != "service error" && !errors.Is(err, tc.wantErr):
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestFeeCalculator_Calculate(t *testing.T) {
	econ := &setup.EconomicConfig{}
	econ.Economics.Betting.BetFees.InitialBetFee = 5
	econ.Economics.Betting.BetFees.BuySharesFee = 2

	calc := feeCalculator{econ: econ}

	tests := []struct {
		name   string
		hasBet bool
		amount int64
		want   betFees
	}{
		{
			name:   "first bet includes initial fee",
			hasBet: false,
			amount: 10,
			want:   betFees{initialFee: 5, transactionFee: 2, totalCost: 17},
		},
		{
			name:   "repeat bet omits initial fee",
			hasBet: true,
			amount: 10,
			want:   betFees{initialFee: 0, transactionFee: 2, totalCost: 12},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fees := calc.Calculate(tc.hasBet, tc.amount)
			if fees != tc.want {
				t.Fatalf("unexpected fees: %+v", fees)
			}
		})
	}
}

func TestBalanceGuard_EnsureSufficient(t *testing.T) {
	guard := balanceGuard{maxDebtAllowed: 50}

	tests := []struct {
		name    string
		balance int64
		cost    int64
		wantErr error
	}{
		{name: "within debt limit", balance: 0, cost: 40},
		{name: "exceeds debt limit", balance: -10, cost: 100, wantErr: ErrInsufficientBalance},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := guard.EnsureSufficient(tc.balance, tc.cost)
			if tc.wantErr == nil && err != nil {
				t.Fatalf("expected success, got %v", err)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

type ledgerRepo struct {
	bet        *models.Bet
	createFunc func(ctx context.Context, bet *models.Bet) error
	hasBetFunc func(ctx context.Context, marketID uint, username string) (bool, error)
}

func newLedgerRepo(opts ...func(*ledgerRepo)) *ledgerRepo {
	repo := &ledgerRepo{
		createFunc: func(context.Context, *models.Bet) error { return nil },
		hasBetFunc: func(context.Context, uint, string) (bool, error) {
			return false, errUnexpectedHelperCall
		},
	}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

func withLedgerCreate(fn func(ctx context.Context, bet *models.Bet) error) func(*ledgerRepo) {
	return func(repo *ledgerRepo) {
		repo.createFunc = fn
	}
}

func withLedgerHasBet(fn func(ctx context.Context, marketID uint, username string) (bool, error)) func(*ledgerRepo) {
	return func(repo *ledgerRepo) {
		repo.hasBetFunc = fn
	}
}

func (l *ledgerRepo) Create(ctx context.Context, bet *models.Bet) error {
	if err := l.createFunc(ctx, bet); err != nil {
		return err
	}
	copyBet := *bet
	l.bet = &copyBet
	return nil
}

func (l *ledgerRepo) UserHasBet(ctx context.Context, marketID uint, username string) (bool, error) {
	return l.hasBetFunc(ctx, marketID, username)
}

type ledgerCall struct {
	username    string
	amount      int64
	transaction string
}

type ledgerUsers struct {
	calls                []ledgerCall
	getUserFunc          func(ctx context.Context, username string) (*dusers.User, error)
	applyTransactionFunc func(ctx context.Context, username string, amount int64, transactionType string) error
}

func newLedgerUsers(opts ...func(*ledgerUsers)) *ledgerUsers {
	users := &ledgerUsers{
		getUserFunc: func(context.Context, string) (*dusers.User, error) {
			return nil, errUnexpectedHelperCall
		},
		applyTransactionFunc: func(context.Context, string, int64, string) error {
			return nil
		},
	}
	for _, opt := range opts {
		opt(users)
	}
	return users
}

func withLedgerUserLookup(fn func(ctx context.Context, username string) (*dusers.User, error)) func(*ledgerUsers) {
	return func(users *ledgerUsers) {
		users.getUserFunc = fn
	}
}

func withLedgerApplyTransaction(fn func(ctx context.Context, username string, amount int64, transactionType string) error) func(*ledgerUsers) {
	return func(users *ledgerUsers) {
		users.applyTransactionFunc = fn
	}
}

func (u *ledgerUsers) GetUser(ctx context.Context, username string) (*dusers.User, error) {
	return u.getUserFunc(ctx, username)
}

func (u *ledgerUsers) ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error {
	if err := u.applyTransactionFunc(ctx, username, amount, transactionType); err != nil {
		return err
	}
	u.calls = append(u.calls, ledgerCall{username: username, amount: amount, transaction: transactionType})
	return nil
}

func TestBetLedger_ChargeAndRecord(t *testing.T) {
	users := newLedgerUsers()
	repo := newLedgerRepo()
	ledger := betLedger{repo: repo, users: users}
	bet := &models.Bet{Username: "bob", Amount: 25}

	if err := ledger.ChargeAndRecord(context.Background(), bet, 25); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(users.calls) != 1 || users.calls[0].transaction != dusers.TransactionBuy || users.calls[0].amount != 25 {
		t.Fatalf("unexpected user calls: %+v", users.calls)
	}
	if repo.bet == nil || repo.bet.Username != bet.Username || repo.bet.Amount != bet.Amount {
		t.Fatalf("expected copied bet persisted, got %+v", repo.bet)
	}
}

func TestBetLedger_ChargeAndRecord_RollsBackOnRepoError(t *testing.T) {
	users := newLedgerUsers()
	repo := newLedgerRepo(withLedgerCreate(func(ctx context.Context, bet *models.Bet) error {
		return errors.New("db down")
	}))
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
	users := newLedgerUsers()
	repo := newLedgerRepo()
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
	tests := []struct {
		name           string
		position       *dmarkets.UserPosition
		sharesOwned    int64
		requested      int64
		want           SaleQuote
		wantErr        error
		wantDustCapErr bool
	}{
		{
			name:        "successful sale",
			position:    pos,
			sharesOwned: 10,
			requested:   23,
			want:        SaleQuote{SharesToSell: 2, SaleValue: 20, Dust: 3},
		},
		{
			name:        "request too small",
			position:    pos,
			sharesOwned: 10,
			requested:   5,
			wantErr:     ErrInvalidAmount,
		},
		{
			name:           "dust cap exceeded",
			position:       pos,
			sharesOwned:    10,
			requested:      35,
			wantDustCapErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := calc.Calculate(tc.position, tc.sharesOwned, tc.requested)
			switch {
			case tc.wantDustCapErr:
				var dustErr ErrDustCapExceeded
				if !errors.As(err, &dustErr) {
					t.Fatalf("expected dust cap error, got %v", err)
				}
			case tc.wantErr != nil:
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
			default:
				if err != nil {
					t.Fatalf("expected success, got %v", err)
				}
				if result != tc.want {
					t.Fatalf("unexpected sale result: %+v", result)
				}
			}
		})
	}
}
