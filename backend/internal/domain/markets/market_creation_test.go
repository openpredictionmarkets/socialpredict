package markets

import (
	"context"
	"errors"
	"testing"
	"time"

	dusers "socialpredict/internal/domain/users"
	dwallet "socialpredict/internal/domain/wallet"
)

type createRepo struct {
	created   *Market
	createErr error
}

func (r *createRepo) Create(_ context.Context, market *Market) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.created = market
	return nil
}

func (r *createRepo) GetByID(context.Context, int64) (*Market, error) {
	panic("unexpected call")
}

func (r *createRepo) UpdateLabels(context.Context, int64, string, string) error {
	panic("unexpected call")
}

func (r *createRepo) List(context.Context, ListFilters) ([]*Market, error) {
	panic("unexpected call")
}

func (r *createRepo) ListByStatus(context.Context, string, Page) ([]*Market, error) {
	panic("unexpected call")
}

func (r *createRepo) Search(context.Context, string, SearchFilters) ([]*Market, error) {
	panic("unexpected call")
}

func (r *createRepo) Delete(context.Context, int64) error {
	panic("unexpected call")
}

func (r *createRepo) ResolveMarket(context.Context, int64, string) error {
	panic("unexpected call")
}

func (r *createRepo) GetUserPosition(context.Context, int64, string) (*UserPosition, error) {
	panic("unexpected call")
}

func (r *createRepo) ListMarketPositions(context.Context, int64) (MarketPositions, error) {
	panic("unexpected call")
}

func (r *createRepo) ListBetsForMarket(context.Context, int64) ([]*Bet, error) {
	panic("unexpected call")
}

func (r *createRepo) CalculatePayoutPositions(context.Context, int64) ([]*PayoutPosition, error) {
	panic("unexpected call")
}

func (r *createRepo) GetPublicMarket(context.Context, int64) (*PublicMarket, error) {
	panic("unexpected call")
}

type createCreatorProfile struct {
	validateErr error
}

func (s createCreatorProfile) ValidateUserExists(context.Context, string) error {
	return s.validateErr
}

func (createCreatorProfile) GetPublicUser(context.Context, string) (*dusers.PublicUser, error) {
	return nil, nil
}

type walletDebitCall struct {
	username string
	amount   int64
	maxDebt  int64
	txType   string
}

type createWallet struct {
	validateCalled bool
	debitCalls     []walletDebitCall
	debitErr       error
}

func (w *createWallet) ValidateBalance(context.Context, string, int64, int64) error {
	w.validateCalled = true
	return nil
}

func (w *createWallet) Debit(_ context.Context, username string, amount int64, maxDebt int64, txType string) error {
	w.debitCalls = append(w.debitCalls, walletDebitCall{
		username: username,
		amount:   amount,
		maxDebt:  maxDebt,
		txType:   txType,
	})
	return w.debitErr
}

func (w *createWallet) Credit(context.Context, string, int64, string) error {
	panic("unexpected call")
}

type createClock struct {
	now time.Time
}

func (c createClock) Now() time.Time {
	return c.now
}

func TestCreateMarket_DebitsWalletWithConfiguredCostAndDebtLimit(t *testing.T) {
	now := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	repo := &createRepo{}
	wallet := &createWallet{}
	service := &Service{
		repo:                  repo,
		creatorProfileService: createCreatorProfile{},
		walletService:         wallet,
		clock:                 createClock{now: now},
		config: Config{
			MinimumFutureHours: 1,
			CreateMarketCost:   75,
			MaximumDebtAllowed: 300,
		},
	}

	market, err := service.CreateMarket(context.Background(), MarketCreateRequest{
		QuestionTitle:      "Will this market debit wallet?",
		Description:        "Create market wallet debit test",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now.Add(3 * time.Hour),
	}, "creator")
	if err != nil {
		t.Fatalf("CreateMarket returned error: %v", err)
	}
	if market == nil {
		t.Fatalf("expected market, got nil")
	}
	if repo.created == nil {
		t.Fatalf("expected repo create to be called")
	}
	if wallet.validateCalled {
		t.Fatalf("did not expect ValidateBalance call")
	}
	if len(wallet.debitCalls) != 1 {
		t.Fatalf("expected 1 wallet debit call, got %d", len(wallet.debitCalls))
	}
	call := wallet.debitCalls[0]
	if call.username != "creator" || call.amount != 75 || call.maxDebt != 300 || call.txType != dwallet.TxFee {
		t.Fatalf("unexpected wallet debit call: %+v", call)
	}
}

func TestCreateMarket_MapsWalletInsufficientBalance(t *testing.T) {
	now := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	repo := &createRepo{}
	wallet := &createWallet{debitErr: dwallet.ErrInsufficientBalance}
	service := &Service{
		repo:                  repo,
		creatorProfileService: createCreatorProfile{},
		walletService:         wallet,
		clock:                 createClock{now: now},
		config: Config{
			MinimumFutureHours: 1,
			CreateMarketCost:   50,
			MaximumDebtAllowed: 100,
		},
	}

	_, err := service.CreateMarket(context.Background(), MarketCreateRequest{
		QuestionTitle:      "Will this fail for insufficient funds?",
		Description:        "insufficient balance mapping test",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now.Add(2 * time.Hour),
	}, "creator")
	if err != ErrInsufficientBalance {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
	if repo.created != nil {
		t.Fatalf("expected no market create when debit fails")
	}
	if wallet.validateCalled {
		t.Fatalf("did not expect ValidateBalance call")
	}
}

func TestCreateMarket_PropagatesNonBalanceWalletError(t *testing.T) {
	now := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	repo := &createRepo{}
	debitErr := errors.New("wallet backend unavailable")
	wallet := &createWallet{debitErr: debitErr}
	service := &Service{
		repo:                  repo,
		creatorProfileService: createCreatorProfile{},
		walletService:         wallet,
		clock:                 createClock{now: now},
		config: Config{
			MinimumFutureHours: 1,
			CreateMarketCost:   50,
			MaximumDebtAllowed: 100,
		},
	}

	_, err := service.CreateMarket(context.Background(), MarketCreateRequest{
		QuestionTitle:      "Will non-balance errors propagate?",
		Description:        "wallet error propagation test",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now.Add(2 * time.Hour),
	}, "creator")
	if !errors.Is(err, debitErr) {
		t.Fatalf("expected original wallet error, got %v", err)
	}
	if repo.created != nil {
		t.Fatalf("expected no market create when debit fails")
	}
	if wallet.validateCalled {
		t.Fatalf("did not expect ValidateBalance call")
	}
}
