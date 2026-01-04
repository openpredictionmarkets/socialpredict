package bets

import (
	"context"
	"strings"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/models"
	"socialpredict/setup"
)

func normalizeOutcome(outcome string) string {
	switch strings.ToUpper(strings.TrimSpace(outcome)) {
	case "YES":
		return "YES"
	case "NO":
		return "NO"
	default:
		return ""
	}
}

func validatePlaceRequest(req PlaceRequest) (string, error) {
	outcome := normalizeOutcome(req.Outcome)
	if outcome == "" {
		return "", ErrInvalidOutcome
	}
	if req.Amount <= 0 {
		return "", ErrInvalidAmount
	}
	return outcome, nil
}

func validateSellRequest(req SellRequest) (string, error) {
	outcome := normalizeOutcome(req.Outcome)
	if outcome == "" {
		return "", ErrInvalidOutcome
	}
	if req.Amount <= 0 {
		return "", ErrInvalidAmount
	}
	return outcome, nil
}

// marketGate ensures markets are open before interacting with them.
type marketGate struct {
	markets MarketService
	clock   Clock
}

func (g marketGate) Open(ctx context.Context, marketID int64) (*dmarkets.Market, error) {
	market, err := g.markets.GetMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if err := ensureMarketOpen(market, g.clock.Now()); err != nil {
		return nil, err
	}
	return market, nil
}

func ensureMarketOpen(market *dmarkets.Market, now time.Time) error {
	if market.Status == "resolved" || now.After(market.ResolutionDateTime) {
		return ErrMarketClosed
	}
	return nil
}

type feeCalculator struct {
	econ *setup.EconomicConfig
}

func (f feeCalculator) Calculate(hasBet bool, amount int64) betFees {
	fees := betFees{
		initialFee:     0,
		transactionFee: int64(f.econ.Economics.Betting.BetFees.BuySharesFee),
	}
	if !hasBet {
		fees.initialFee = int64(f.econ.Economics.Betting.BetFees.InitialBetFee)
	}
	fees.totalCost = amount + fees.initialFee + fees.transactionFee
	return fees
}

type balanceGuard struct {
	maxDebtAllowed int64
}

func (g balanceGuard) EnsureSufficient(balance, totalCost int64) error {
	if balance-totalCost < -g.maxDebtAllowed {
		return ErrInsufficientBalance
	}
	return nil
}

type betLedger struct {
	repo  Repository
	users UserService
}

func (l betLedger) ChargeAndRecord(ctx context.Context, bet *models.Bet, totalCost int64) error {
	if err := l.users.ApplyTransaction(ctx, bet.Username, totalCost, dusers.TransactionBuy); err != nil {
		return err
	}

	if err := l.repo.Create(ctx, bet); err != nil {
		_ = l.users.ApplyTransaction(ctx, bet.Username, totalCost, dusers.TransactionRefund)
		return err
	}
	return nil
}

func (l betLedger) CreditSale(ctx context.Context, bet *models.Bet, saleValue int64) error {
	if err := l.users.ApplyTransaction(ctx, bet.Username, saleValue, dusers.TransactionSale); err != nil {
		return err
	}
	if err := l.repo.Create(ctx, bet); err != nil {
		_ = l.users.ApplyTransaction(ctx, bet.Username, saleValue, dusers.TransactionBuy)
		return err
	}
	return nil
}
