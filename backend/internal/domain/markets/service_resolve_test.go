package markets_test

import (
	"context"
	"errors"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	dwallet "socialpredict/internal/domain/wallet"
)

type resolveRepo struct {
	market       *markets.Market
	bets         []*markets.Bet
	positions    []*markets.PayoutPosition
	resolveErr   error
	listBetsErr  error
	payoutPosErr error
}

func (r *resolveRepo) Create(context.Context, *markets.Market) error { panic("unexpected call") }
func (r *resolveRepo) UpdateLabels(context.Context, int64, string, string) error {
	panic("unexpected call")
}
func (r *resolveRepo) List(context.Context, markets.ListFilters) ([]*markets.Market, error) {
	panic("unexpected call")
}
func (r *resolveRepo) ListByStatus(context.Context, string, markets.Page) ([]*markets.Market, error) {
	panic("unexpected call")
}
func (r *resolveRepo) Search(context.Context, string, markets.SearchFilters) ([]*markets.Market, error) {
	panic("unexpected call")
}
func (r *resolveRepo) Delete(context.Context, int64) error { panic("unexpected call") }

func (r *resolveRepo) GetByID(context.Context, int64) (*markets.Market, error) {
	if r.market == nil {
		return nil, markets.ErrMarketNotFound
	}
	return r.market, nil
}

func (r *resolveRepo) ResolveMarket(context.Context, int64, string) error {
	if r.resolveErr != nil {
		return r.resolveErr
	}
	if r.market != nil {
		r.market.Status = "resolved"
	}
	return nil
}

func (r *resolveRepo) GetUserPosition(context.Context, int64, string) (*markets.UserPosition, error) {
	panic("unexpected call")
}

func (r *resolveRepo) ListMarketPositions(context.Context, int64) (markets.MarketPositions, error) {
	panic("unexpected call")
}

func (r *resolveRepo) ListBetsForMarket(context.Context, int64) ([]*markets.Bet, error) {
	if r.listBetsErr != nil {
		return nil, r.listBetsErr
	}
	return r.bets, nil
}

func (r *resolveRepo) CalculatePayoutPositions(context.Context, int64) ([]*markets.PayoutPosition, error) {
	if r.payoutPosErr != nil {
		return nil, r.payoutPosErr
	}
	return r.positions, nil
}

func (r *resolveRepo) GetPublicMarket(context.Context, int64) (*markets.PublicMarket, error) {
	return nil, nil
}

type resolveUserService struct {
	applied []struct {
		username string
		amount   int64
		txType   string
	}
}

func (resolveUserService) ValidateUserExists(context.Context, string) error { return nil }
func (resolveUserService) ValidateUserBalance(context.Context, string, int64, int64) error {
	return nil
}
func (resolveUserService) DeductBalance(context.Context, string, int64) error { return nil }
func (s *resolveUserService) ApplyTransaction(ctx context.Context, username string, amount int64, tx string) error {
	s.applied = append(s.applied, struct {
		username string
		amount   int64
		txType   string
	}{username: username, amount: amount, txType: tx})
	return nil
}

func (resolveUserService) GetPublicUser(context.Context, string) (*dusers.PublicUser, error) {
	return nil, nil
}

type nopClock struct{}

func (nopClock) Now() time.Time { return time.Now() }

func TestResolveMarketRefundsOnNA(t *testing.T) {
	repo := &resolveRepo{
		market: &markets.Market{
			ID:              1,
			CreatorUsername: "creator",
			Status:          "active",
		},
		bets: []*markets.Bet{
			{Username: "alice", Amount: 50},
			{Username: "bob", Amount: 30},
		},
	}
	userSvc := &resolveUserService{}
	service := markets.NewService(repo, userSvc, nopClock{}, markets.Config{})

	if err := service.ResolveMarket(context.Background(), 1, "N/A", "creator"); err != nil {
		t.Fatalf("ResolveMarket returned error: %v", err)
	}

	if len(userSvc.applied) != 2 {
		t.Fatalf("expected 2 refund transactions, got %d", len(userSvc.applied))
	}

	for _, call := range userSvc.applied {
		if call.txType != dwallet.TxRefund {
			t.Fatalf("expected refund transaction, got %s", call.txType)
		}
	}
}

func TestResolveMarketPaysWinners(t *testing.T) {
	repo := &resolveRepo{
		market: &markets.Market{
			ID:              42,
			CreatorUsername: "creator",
			Status:          "active",
		},
		positions: []*markets.PayoutPosition{
			{Username: "winner", Value: 120},
			{Username: "loser", Value: 0},
		},
	}
	userSvc := &resolveUserService{}
	service := markets.NewService(repo, userSvc, nopClock{}, markets.Config{})

	if err := service.ResolveMarket(context.Background(), 42, "YES", "creator"); err != nil {
		t.Fatalf("ResolveMarket returned error: %v", err)
	}

	if len(userSvc.applied) != 1 {
		t.Fatalf("expected single payout, got %d", len(userSvc.applied))
	}

	call := userSvc.applied[0]
	if call.username != "winner" || call.amount != 120 || call.txType != dwallet.TxWin {
		t.Fatalf("unexpected payout %+v", call)
	}
}

type resolveCreatorProfile struct{}

func (resolveCreatorProfile) ValidateUserExists(context.Context, string) error { return nil }
func (resolveCreatorProfile) GetPublicUser(context.Context, string) (*dusers.PublicUser, error) {
	return nil, nil
}

type resolveWallet struct {
	creditErr  error
	failOnCall int // 1-indexed; 0 = fail all calls when creditErr is set
	calls      int
	credited   []struct {
		username string
		amount   int64
		txType   string
	}
}

func (w *resolveWallet) ValidateBalance(context.Context, string, int64, int64) error { return nil }
func (w *resolveWallet) Debit(context.Context, string, int64, int64, string) error   { return nil }

func (w *resolveWallet) Credit(_ context.Context, username string, amount int64, txType string) error {
	w.calls++
	if w.creditErr != nil && (w.failOnCall == 0 || w.calls == w.failOnCall) {
		return w.creditErr
	}
	w.credited = append(w.credited, struct {
		username string
		amount   int64
		txType   string
	}{username, amount, txType})
	return nil
}

func TestResolveMarket_RefundCreditFailureMidLoop(t *testing.T) {
	repo := &resolveRepo{
		market: &markets.Market{
			ID:              1,
			CreatorUsername: "creator",
			Status:          "active",
		},
		bets: []*markets.Bet{
			{Username: "alice", Amount: 50},
			{Username: "bob", Amount: 30},
		},
	}
	wallet := &resolveWallet{creditErr: errors.New("wallet down"), failOnCall: 2}
	service := markets.NewServiceWithWallet(repo, resolveCreatorProfile{}, wallet, nopClock{}, markets.Config{})

	err := service.ResolveMarket(context.Background(), 1, "N/A", "creator")
	if err == nil {
		t.Fatalf("expected error from credit failure")
	}
	// First refund succeeded, second failed — partial refund state
	if len(wallet.credited) != 1 {
		t.Fatalf("expected 1 successful credit before failure, got %d", len(wallet.credited))
	}
	if wallet.credited[0].username != "alice" || wallet.credited[0].amount != 50 {
		t.Fatalf("unexpected first credit: %+v", wallet.credited[0])
	}
}

func TestResolveMarket_PayoutCreditFailureMidLoop(t *testing.T) {
	repo := &resolveRepo{
		market: &markets.Market{
			ID:              42,
			CreatorUsername: "creator",
			Status:          "active",
		},
		positions: []*markets.PayoutPosition{
			{Username: "winner1", Value: 120},
			{Username: "winner2", Value: 80},
		},
	}
	wallet := &resolveWallet{creditErr: errors.New("wallet down"), failOnCall: 2}
	service := markets.NewServiceWithWallet(repo, resolveCreatorProfile{}, wallet, nopClock{}, markets.Config{})

	err := service.ResolveMarket(context.Background(), 42, "YES", "creator")
	if err == nil {
		t.Fatalf("expected error from credit failure")
	}
	// First payout succeeded, second failed — partial payout state
	if len(wallet.credited) != 1 {
		t.Fatalf("expected 1 successful credit before failure, got %d", len(wallet.credited))
	}
	if wallet.credited[0].username != "winner1" || wallet.credited[0].amount != 120 {
		t.Fatalf("unexpected first credit: %+v", wallet.credited[0])
	}
}

func TestResolveMarket_ListBetsFailure(t *testing.T) {
	repo := &resolveRepo{
		market: &markets.Market{
			ID:              1,
			CreatorUsername: "creator",
			Status:          "active",
		},
		listBetsErr: errors.New("db connection lost"),
	}
	wallet := &resolveWallet{}
	service := markets.NewServiceWithWallet(repo, resolveCreatorProfile{}, wallet, nopClock{}, markets.Config{})

	err := service.ResolveMarket(context.Background(), 1, "N/A", "creator")
	if err == nil {
		t.Fatalf("expected error from ListBetsForMarket failure")
	}
	if wallet.calls != 0 {
		t.Fatalf("expected no wallet calls when repo fails, got %d", wallet.calls)
	}
}

func TestResolveMarket_CalculatePayoutPositionsFailure(t *testing.T) {
	repo := &resolveRepo{
		market: &markets.Market{
			ID:              42,
			CreatorUsername: "creator",
			Status:          "active",
		},
		payoutPosErr: errors.New("db connection lost"),
	}
	wallet := &resolveWallet{}
	service := markets.NewServiceWithWallet(repo, resolveCreatorProfile{}, wallet, nopClock{}, markets.Config{})

	err := service.ResolveMarket(context.Background(), 42, "YES", "creator")
	if err == nil {
		t.Fatalf("expected error from CalculatePayoutPositions failure")
	}
	if wallet.calls != 0 {
		t.Fatalf("expected no wallet calls when repo fails, got %d", wallet.calls)
	}
}

func TestResolveMarketRejectsUnauthorized(t *testing.T) {
	repo := &resolveRepo{
		market: &markets.Market{
			ID:              5,
			CreatorUsername: "owner",
			Status:          "active",
		},
	}
	userSvc := &resolveUserService{}
	service := markets.NewService(repo, userSvc, nopClock{}, markets.Config{})

	err := service.ResolveMarket(context.Background(), 5, "YES", "intruder")
	if err != markets.ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}
