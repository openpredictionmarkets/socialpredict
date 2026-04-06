package analytics

import (
	"context"
	"errors"
)

// ComputeUserFinancials calculates comprehensive financial metrics for a user.
func (s *Service) ComputeUserFinancials(ctx context.Context, req FinancialSnapshotRequest) (*FinancialSnapshot, error) {
	if req.Username == "" {
		return nil, errors.New("username is required")
	}

	if s.repo == nil {
		return nil, errors.New("repository not provided")
	}

	if s.econLoader == nil {
		return nil, errors.New("economic configuration loader not provided")
	}

	positions, err := s.repo.UserMarketPositions(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	econConfig := s.econLoader()

	snapshot := &FinancialSnapshot{
		AccountBalance:     req.AccountBalance,
		MaximumDebtAllowed: econConfig.Economics.User.MaximumDebtAllowed,
	}

	for _, pos := range positions {
		profit := pos.Value - pos.TotalSpent

		snapshot.AmountInPlay += pos.Value
		snapshot.TotalSpent += pos.TotalSpent
		snapshot.TradingProfits += profit

		if pos.IsResolved {
			snapshot.RealizedProfits += profit
			snapshot.RealizedValue += pos.Value
		} else {
			snapshot.PotentialProfits += profit
			snapshot.PotentialValue += pos.Value
			snapshot.AmountInPlayActive += pos.Value
			snapshot.TotalSpentInPlay += pos.TotalSpentInPlay
		}
	}

	if req.AccountBalance < 0 {
		snapshot.AmountBorrowed = -req.AccountBalance
	}

	snapshot.RetainedEarnings = req.AccountBalance - snapshot.AmountInPlay
	snapshot.Equity = snapshot.RetainedEarnings + snapshot.AmountInPlay - snapshot.AmountBorrowed
	snapshot.TotalProfits = snapshot.TradingProfits + snapshot.WorkProfits

	return snapshot, nil
}
