package bets

import (
	"context"
	"errors"
	"testing"
	"time"

	"socialpredict/internal/domain/boundary"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
)

var errUnexpectedHelperCall = errors.New("unexpected call")

type stubMarketService struct {
	marketGetter   stubMarketGetter
	positionGetter stubPositionGetter
}

func newStubMarketService(opts ...func(*stubMarketService)) stubMarketService {
	stub := stubMarketService{
		marketGetter: stubMarketGetter{
			getMarketFunc: func(context.Context, int64) (*dmarkets.Market, error) {
				return nil, errUnexpectedHelperCall
			},
		},
		positionGetter: stubPositionGetter{
			getUserPositionInMarketFn: func(context.Context, int64, string) (*dmarkets.UserPosition, error) {
				return nil, errUnexpectedHelperCall
			},
		},
	}
	for _, opt := range opts {
		opt(&stub)
	}
	return stub
}

func withStubMarket(getMarket func(ctx context.Context, id int64) (*dmarkets.Market, error)) func(*stubMarketService) {
	return func(stub *stubMarketService) {
		stub.marketGetter.getMarketFunc = getMarket
	}
}

func withStubPosition(getPosition func(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error)) func(*stubMarketService) {
	return func(stub *stubMarketService) {
		stub.positionGetter.getUserPositionInMarketFn = getPosition
	}
}

type stubMarketGetter struct {
	getMarketFunc func(ctx context.Context, id int64) (*dmarkets.Market, error)
}

func (s stubMarketService) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	return s.marketGetter.GetMarket(ctx, id)
}

type stubPositionGetter struct {
	getUserPositionInMarketFn func(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error)
}

func (s stubMarketService) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	return s.positionGetter.GetUserPositionInMarket(ctx, marketID, username)
}

func (s stubMarketGetter) GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error) {
	if s.getMarketFunc == nil {
		return nil, errUnexpectedHelperCall
	}
	return s.getMarketFunc(ctx, id)
}

func (s stubPositionGetter) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*dmarkets.UserPosition, error) {
	if s.getUserPositionInMarketFn == nil {
		return nil, errUnexpectedHelperCall
	}
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
		return helperTestTime()
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

	t.Run("zero value clock remains substitutable", func(t *testing.T) {
		gate := marketGate{
			markets: newStubMarketService(withStubMarket(func(context.Context, int64) (*dmarkets.Market, error) {
				return &dmarkets.Market{Status: "active", ResolutionDateTime: helperTestTime().Add(time.Hour)}, nil
			})),
			clock: gateClock{},
		}
		if _, err := gate.Open(context.Background(), 1); err != nil {
			t.Fatalf("expected zero-value clock to remain usable, got %v", err)
		}
	})
}

func TestFeeCalculator_Calculate(t *testing.T) {
	calc := feeCalculator{config: Config{
		InitialBetFee: 5,
		BuySharesFee:  2,
	}}

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

	if zeroFees := (feeCalculator{config: Config{}}).Calculate(false, 0); zeroFees.totalCost != 0 {
		t.Fatalf("expected zero-value config to remain usable, got %+v", zeroFees)
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

	if err := (balanceGuard{}).EnsureSufficient(0, 0); err != nil {
		t.Fatalf("expected zero-value guard to allow no-op cost, got %v", err)
	}
}

type ledgerRepo struct {
	bet     *boundary.Bet
	creator ledgerBetCreator
	checker ledgerBetChecker
}

func newLedgerRepo(opts ...func(*ledgerRepo)) *ledgerRepo {
	repo := &ledgerRepo{
		creator: ledgerBetCreator{
			createFunc: func(context.Context, *boundary.Bet) error { return nil },
		},
		checker: ledgerBetChecker{
			hasBetFunc: func(context.Context, uint, string) (bool, error) {
				return false, errUnexpectedHelperCall
			},
		},
	}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

func withLedgerCreate(fn func(ctx context.Context, bet *boundary.Bet) error) func(*ledgerRepo) {
	return func(repo *ledgerRepo) {
		repo.creator.createFunc = fn
	}
}

func withLedgerHasBet(fn func(ctx context.Context, marketID uint, username string) (bool, error)) func(*ledgerRepo) {
	return func(repo *ledgerRepo) {
		repo.checker.hasBetFunc = fn
	}
}

type ledgerBetCreator struct {
	createFunc func(ctx context.Context, bet *boundary.Bet) error
}

func (l *ledgerRepo) Create(ctx context.Context, bet *boundary.Bet) error {
	return l.creator.Create(ctx, bet, l)
}

type ledgerBetChecker struct {
	hasBetFunc func(ctx context.Context, marketID uint, username string) (bool, error)
}

func (l *ledgerRepo) UserHasBet(ctx context.Context, marketID uint, username string) (bool, error) {
	return l.checker.UserHasBet(ctx, marketID, username)
}

func (c ledgerBetCreator) Create(ctx context.Context, bet *boundary.Bet, repo *ledgerRepo) error {
	if c.createFunc == nil {
		return errUnexpectedHelperCall
	}
	if err := c.createFunc(ctx, bet); err != nil {
		return err
	}
	if bet == nil {
		repo.bet = nil
		return nil
	}
	copyBet := *bet
	repo.bet = &copyBet
	return nil
}

func (c ledgerBetChecker) UserHasBet(ctx context.Context, marketID uint, username string) (bool, error) {
	if c.hasBetFunc == nil {
		return false, errUnexpectedHelperCall
	}
	return c.hasBetFunc(ctx, marketID, username)
}

type ledgerCall struct {
	username    string
	amount      int64
	transaction string
}

type ledgerUsers struct {
	calls   []ledgerCall
	reader  ledgerUserReader
	applier ledgerTransactionApplier
}

func newLedgerUsers(opts ...func(*ledgerUsers)) *ledgerUsers {
	users := &ledgerUsers{
		reader: ledgerUserReader{
			getUserFunc: func(context.Context, string) (*dusers.User, error) {
				return nil, errUnexpectedHelperCall
			},
		},
		applier: ledgerTransactionApplier{
			applyTransactionFunc: func(context.Context, string, int64, string) error {
				return nil
			},
		},
	}
	for _, opt := range opts {
		opt(users)
	}
	return users
}

func withLedgerUserLookup(fn func(ctx context.Context, username string) (*dusers.User, error)) func(*ledgerUsers) {
	return func(users *ledgerUsers) {
		users.reader.getUserFunc = fn
	}
}

func withLedgerApplyTransaction(fn func(ctx context.Context, username string, amount int64, transactionType string) error) func(*ledgerUsers) {
	return func(users *ledgerUsers) {
		users.applier.applyTransactionFunc = fn
	}
}

type ledgerUserReader struct {
	getUserFunc func(ctx context.Context, username string) (*dusers.User, error)
}

func (u *ledgerUsers) GetUser(ctx context.Context, username string) (*dusers.User, error) {
	return u.reader.GetUser(ctx, username)
}

type ledgerTransactionApplier struct {
	applyTransactionFunc func(ctx context.Context, username string, amount int64, transactionType string) error
}

func (u *ledgerUsers) ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error {
	return u.applier.ApplyTransaction(ctx, username, amount, transactionType, &u.calls)
}

func (u ledgerUserReader) GetUser(ctx context.Context, username string) (*dusers.User, error) {
	if u.getUserFunc == nil {
		return nil, errUnexpectedHelperCall
	}
	return u.getUserFunc(ctx, username)
}

func (u ledgerTransactionApplier) ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string, calls *[]ledgerCall) error {
	if u.applyTransactionFunc == nil {
		return errUnexpectedHelperCall
	}
	if err := u.applyTransactionFunc(ctx, username, amount, transactionType); err != nil {
		return err
	}
	*calls = append(*calls, ledgerCall{username: username, amount: amount, transaction: transactionType})
	return nil
}

func TestBetLedger_ChargeAndRecord(t *testing.T) {
	users := newLedgerUsers()
	repo := newLedgerRepo()
	ledger := betLedger{repo: repo, users: users}
	bet := &boundary.Bet{Username: "bob", Amount: 25}

	if err := ledger.ChargeAndRecord(context.Background(), bet, 25); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(users.calls) != 1 || users.calls[0].transaction != dusers.TransactionBuy || users.calls[0].amount != 25 {
		t.Fatalf("unexpected user calls: %+v", users.calls)
	}
	if repo.bet == nil || repo.bet.Username != bet.Username || repo.bet.Amount != bet.Amount {
		t.Fatalf("expected copied bet persisted, got %+v", repo.bet)
	}

	if _, err := (&ledgerRepo{}).UserHasBet(context.Background(), 1, "bob"); !errors.Is(err, errUnexpectedHelperCall) {
		t.Fatalf("expected zero-value repo to fail predictably, got %v", err)
	}
}

func TestBetLedger_ChargeAndRecord_RollsBackOnRepoError(t *testing.T) {
	users := newLedgerUsers()
	repo := newLedgerRepo(withLedgerCreate(func(ctx context.Context, bet *boundary.Bet) error {
		return errors.New("db down")
	}))
	ledger := betLedger{repo: repo, users: users}

	err := ledger.ChargeAndRecord(context.Background(), &boundary.Bet{Username: "alice"}, 10)
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

	if err := ledger.CreditSale(context.Background(), &boundary.Bet{Username: "alice"}, 15); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(users.calls) != 1 || users.calls[0].transaction != dusers.TransactionSale || users.calls[0].amount != 15 {
		t.Fatalf("unexpected user calls: %+v", users.calls)
	}

	if _, err := (&ledgerUsers{}).GetUser(context.Background(), "alice"); !errors.Is(err, errUnexpectedHelperCall) {
		t.Fatalf("expected zero-value users to fail predictably, got %v", err)
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

	if _, err := (saleCalculator{}).Calculate(nil, 0, 1); !errors.Is(err, ErrNoPosition) {
		t.Fatalf("expected zero-value calculator to keep base contract, got %v", err)
	}
}
