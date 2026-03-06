package bets

import (
	"context"

	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"
)

// Sell processes a sell request for credits.
func (s *Service) Sell(ctx context.Context, req SellRequest) (*SellResult, error) {
	outcome, err := s.sellValidator.Validate(ctx, req)
	if err != nil {
		return nil, err
	}
	if outcome == "" {
		return nil, ErrInvalidOutcome
	}

	if _, err := s.marketGate.Open(ctx, int64(req.MarketID)); err != nil {
		return nil, err
	}

	sharesOwned, position, err := s.loadUserShares(ctx, req, outcome)
	if err != nil {
		return nil, err
	}

	sale, err := s.saleCalculator.Calculate(position, sharesOwned, req.Amount)
	if err != nil {
		return nil, err
	}
	if sale.SharesToSell == 0 {
		return nil, ErrInsufficientShares
	}

	now := s.clock.Now()
	bet := &models.Bet{
		Username: req.Username,
		MarketID: req.MarketID,
		Amount:   -sale.SharesToSell,
		Outcome:  outcome,
		PlacedAt: now,
	}
	if err := s.ledger.CreditSale(ctx, bet, sale.SaleValue); err != nil {
		return nil, err
	}

	return &SellResult{
		Username:      req.Username,
		MarketID:      req.MarketID,
		SharesSold:    sale.SharesToSell,
		SaleValue:     sale.SaleValue,
		Dust:          sale.Dust,
		Outcome:       outcome,
		TransactionAt: now,
	}, nil
}

func (s *Service) loadUserShares(ctx context.Context, req SellRequest, outcome string) (int64, *dmarkets.UserPosition, error) {
	position, err := s.markets.GetUserPositionInMarket(ctx, int64(req.MarketID), req.Username)
	if err != nil {
		return 0, nil, err
	}
	if position == nil {
		return 0, nil, ErrNoPosition
	}

	sharesOwned, err := sharesOwnedForOutcome(position, outcome)
	if err != nil {
		return 0, nil, err
	}

	return sharesOwned, position, nil
}

// SaleQuote summarises how a sale request would be executed.
// Exported so alternative SaleCalculator implementations can return it.
type SaleQuote struct {
	SharesToSell int64
	SaleValue    int64
	Dust         int64
}

type saleCalculator struct {
	maxDustPerSale int64
}

func (s saleCalculator) Calculate(pos *dmarkets.UserPosition, sharesOwned int64, creditsRequested int64) (SaleQuote, error) {
	if pos == nil {
		return SaleQuote{}, ErrNoPosition
	}
	if sharesOwned <= 0 {
		return SaleQuote{}, ErrNoPosition
	}
	if err := validatePositionValue(pos.Value); err != nil {
		return SaleQuote{}, err
	}

	valuePerShare := pos.Value / sharesOwned
	if valuePerShare <= 0 {
		return SaleQuote{}, ErrNoPosition
	}
	if creditsRequested < valuePerShare {
		return SaleQuote{}, ErrInvalidAmount
	}

	sharesToSell := creditsRequested / valuePerShare
	if sharesToSell > sharesOwned {
		sharesToSell = sharesOwned
	}
	if sharesToSell == 0 {
		return SaleQuote{}, ErrInsufficientShares
	}

	saleValue := sharesToSell * valuePerShare
	dust := calculateDust(creditsRequested, saleValue)

	if err := validateDustCap(dust, s.maxDustPerSale); err != nil {
		return SaleQuote{}, err
	}

	return SaleQuote{SharesToSell: sharesToSell, SaleValue: saleValue, Dust: dust}, nil
}

func sharesOwnedForOutcome(pos *dmarkets.UserPosition, outcome string) (int64, error) {
	switch outcome {
	case "YES":
		if pos.YesSharesOwned == 0 {
			return 0, ErrNoPosition
		}
		return pos.YesSharesOwned, nil
	case "NO":
		if pos.NoSharesOwned == 0 {
			return 0, ErrNoPosition
		}
		return pos.NoSharesOwned, nil
	default:
		return 0, ErrInvalidOutcome
	}
}

func validatePositionValue(value int64) error {
	if value <= 0 {
		return ErrNoPosition
	}
	return nil
}

func calculateDust(requested, saleValue int64) int64 {
	dust := requested - saleValue
	if dust < 0 {
		return 0
	}
	return dust
}

func validateDustCap(dust int64, cap int64) error {
	if cap > 0 && dust > cap {
		return ErrDustCapExceeded{Cap: cap, Requested: dust}
	}
	return nil
}
