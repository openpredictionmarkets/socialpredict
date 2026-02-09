package bets

import (
	"context"
	"errors"

	dwallet "socialpredict/internal/domain/wallet"
	"socialpredict/models"
)

// Place creates a buy bet after validating market status and user balance.
func (s *Service) Place(ctx context.Context, req PlaceRequest) (*PlacedBet, error) {
	outcome, err := validatePlaceRequest(req)
	if err != nil {
		return nil, err
	}

	if _, err := s.marketGate.Open(ctx, int64(req.MarketID)); err != nil {
		return nil, err
	}

	hasBet, err := s.loadBetStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	fees := s.fees.Calculate(hasBet, req.Amount)
	if err := s.wallet.ValidateBalance(ctx, req.Username, fees.totalCost, s.balances.maxDebtAllowed); err != nil {
		if errors.Is(err, dwallet.ErrInsufficientBalance) {
			return nil, ErrInsufficientBalance
		}
		return nil, err
	}

	now := s.clock.Now()
	bet := &models.Bet{
		Username: req.Username,
		MarketID: req.MarketID,
		Amount:   req.Amount,
		Outcome:  outcome,
		PlacedAt: now,
	}

	if err := s.ledger.ChargeAndRecord(ctx, bet, fees.totalCost); err != nil {
		return nil, err
	}

	return placedBetFromModel(bet), nil
}

func (s *Service) loadBetStatus(ctx context.Context, req PlaceRequest) (bool, error) {
	hasBet, err := s.repo.UserHasBet(ctx, req.MarketID, req.Username)
	if err != nil {
		return false, err
	}

	return hasBet, nil
}

type betFees struct {
	initialFee     int64
	transactionFee int64
	totalCost      int64
}

func placedBetFromModel(bet *models.Bet) *PlacedBet {
	return &PlacedBet{
		Username: bet.Username,
		MarketID: bet.MarketID,
		Amount:   bet.Amount,
		Outcome:  bet.Outcome,
		PlacedAt: bet.PlacedAt,
	}
}
