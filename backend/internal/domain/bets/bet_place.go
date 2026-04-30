package bets

import (
	"context"

	dusers "socialpredict/internal/domain/users"
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

	if s.placeUnit == nil {
		return nil, ErrPlaceTransactionUnavailable
	}

	return s.placeInTransaction(ctx, req, outcome)
}

func (s *Service) placeInTransaction(ctx context.Context, req PlaceRequest, outcome string) (*PlacedBet, error) {
	var placed *PlacedBet
	err := s.placeUnit.PlaceBetTransaction(ctx, func(txCtx context.Context, repo Repository, users UserService) error {
		user, hasBet, err := s.loadUserAndBetStatus(txCtx, repo, users, req)
		if err != nil {
			return err
		}

		fees := s.fees.Calculate(hasBet, req.Amount)
		if err := s.balances.EnsureSufficient(user.AccountBalance, fees.totalCost); err != nil {
			return err
		}

		bet := req.NewBet(outcome, s.clock.Now())
		if err := users.ApplyTransaction(txCtx, bet.Username, fees.totalCost, dusers.TransactionBuy); err != nil {
			return err
		}
		if err := repo.Create(txCtx, bet); err != nil {
			return err
		}
		placed = new(PlacedBet).FromModel(bet)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return placed, nil
}

func (s *Service) loadUserAndBetStatus(ctx context.Context, repo Repository, users UserService, req PlaceRequest) (*dusers.User, bool, error) {
	user, err := users.GetUser(ctx, req.Username)
	if err != nil {
		return nil, false, err
	}
	if user == nil {
		return nil, false, dusers.ErrUserNotFound
	}

	hasBet, err := repo.UserHasBet(ctx, req.MarketID, req.Username)
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
