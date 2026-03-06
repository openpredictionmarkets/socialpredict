package bets

import (
	"context"

	dusers "socialpredict/internal/domain/users"
	"socialpredict/models"
)

// Place creates a buy bet after validating market status and user balance.
func (s *Service) Place(ctx context.Context, req PlaceRequest) (*PlacedBet, error) {
	outcome, err := s.placeValidator.Validate(ctx, req)
	if err != nil {
		return nil, err
	}
	if outcome == "" {
		return nil, ErrInvalidOutcome
	}

	if _, err := s.marketGate.Open(ctx, int64(req.MarketID)); err != nil {
		return nil, err
	}

	user, hasBet, err := s.loadUserAndBetStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	fees := s.fees.Calculate(hasBet, req.Amount)
	if err := s.balances.EnsureSufficient(user.AccountBalance, fees.totalCost); err != nil {
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

func (s *Service) loadUserAndBetStatus(ctx context.Context, req PlaceRequest) (*dusers.User, bool, error) {
	user, err := s.users.GetUser(ctx, req.Username)
	if err != nil {
		return nil, false, err
	}
	if user == nil {
		return nil, false, dusers.ErrUserNotFound
	}

	hasBet, err := s.repo.UserHasBet(ctx, req.MarketID, req.Username)
	if err != nil {
		return nil, false, err
	}

	return user, hasBet, nil
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
