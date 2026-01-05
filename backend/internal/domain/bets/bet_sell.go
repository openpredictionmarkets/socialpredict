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
	if sale.sharesToSell == 0 {
		return nil, ErrInsufficientShares
	}

	now := s.clock.Now()
	bet := &models.Bet{
		Username: req.Username,
		MarketID: req.MarketID,
		Amount:   -sale.sharesToSell,
		Outcome:  outcome,
		PlacedAt: now,
	}
	if err := s.ledger.CreditSale(ctx, bet, sale.saleValue); err != nil {
		return nil, err
	}

	return &SellResult{
		Username:      req.Username,
		MarketID:      req.MarketID,
		SharesSold:    sale.sharesToSell,
		SaleValue:     sale.saleValue,
		Dust:          sale.dust,
		Outcome:       outcome,
		TransactionAt: now,
	}, nil
}

func (s *Service) loadUserShares(ctx context.Context, req SellRequest, outcome string) (int64, *dmarkets.UserPosition, error) {
	position, err := s.markets.GetUserPositionInMarket(ctx, int64(req.MarketID), req.Username)
	if err != nil {
		return 0, nil, err
	}

	sharesOwned, err := sharesOwnedForOutcome(position, outcome)
	if err != nil {
		return 0, nil, err
	}

	return sharesOwned, position, nil
}

type saleResult struct {
	sharesToSell int64
	saleValue    int64
	dust         int64
}

type saleCalculator struct {
	maxDustPerSale int64
}

func (s saleCalculator) Calculate(pos *dmarkets.UserPosition, sharesOwned int64, creditsRequested int64) (saleResult, error) {
	if err := validatePositionValue(pos.Value); err != nil {
		return saleResult{}, err
	}

	valuePerShare := pos.Value / sharesOwned
	if valuePerShare <= 0 {
		return saleResult{}, ErrNoPosition
	}
	if creditsRequested < valuePerShare {
		return saleResult{}, ErrInvalidAmount
	}

	sharesToSell := creditsRequested / valuePerShare
	if sharesToSell > sharesOwned {
		sharesToSell = sharesOwned
	}
	if sharesToSell == 0 {
		return saleResult{}, ErrInsufficientShares
	}

	saleValue := sharesToSell * valuePerShare
	dust := calculateDust(creditsRequested, saleValue)

	if err := validateDustCap(dust, s.maxDustPerSale); err != nil {
		return saleResult{}, err
	}

	return saleResult{sharesToSell: sharesToSell, saleValue: saleValue, dust: dust}, nil
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
