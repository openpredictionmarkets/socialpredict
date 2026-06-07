package analytics

import (
	"context"
	"errors"
	"time"

	positionsmath "socialpredict/internal/domain/math/positions"
)

// ComputeUserFinancials calculates comprehensive financial metrics for a user.
func (s *Service) ComputeUserFinancials(ctx context.Context, req FinancialSnapshotRequest) (*FinancialSnapshot, error) {
	if req.Username == "" {
		return nil, errors.New("username is required")
	}

	if s.repo == nil {
		return nil, errors.New("repository not provided")
	}

	positions, err := s.repo.UserMarketPositions(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	snapshot := NewUserFinancialMetricSnapshotCalculator(s.config).Calculate(req, positions, time.Time{})
	return &snapshot.Financial, nil
}

// UserFinancialMetricSnapshotCalculator calculates display-only user financial
// read models from canonical position data.
type UserFinancialMetricSnapshotCalculator struct {
	config Config
}

func NewUserFinancialMetricSnapshotCalculator(config Config) UserFinancialMetricSnapshotCalculator {
	return UserFinancialMetricSnapshotCalculator{config: config}
}

func (c UserFinancialMetricSnapshotCalculator) Calculate(req FinancialSnapshotRequest, positions []positionsmath.MarketPosition, generatedAt time.Time) UserFinancialMetricSnapshot {
	financial := FinancialSnapshot{
		AccountBalance:     req.AccountBalance,
		MaximumDebtAllowed: c.config.MaximumDebtAllowed,
	}
	for _, pos := range positions {
		profit := pos.Value - pos.TotalSpent

		financial.AmountInPlay += pos.Value
		financial.TotalSpent += pos.TotalSpent
		financial.TradingProfits += profit

		if pos.IsResolved {
			financial.RealizedProfits += profit
			financial.RealizedValue += pos.Value
		} else {
			financial.PotentialProfits += profit
			financial.PotentialValue += pos.Value
			financial.AmountInPlayActive += pos.Value
			financial.TotalSpentInPlay += pos.TotalSpentInPlay
		}
	}

	if req.AccountBalance < 0 {
		financial.AmountBorrowed = -req.AccountBalance
	}

	financial.RetainedEarnings = req.AccountBalance - financial.AmountInPlay
	financial.Equity = financial.RetainedEarnings + financial.AmountInPlay - financial.AmountBorrowed
	financial.TotalProfits = financial.TradingProfits + financial.WorkProfits

	return UserFinancialMetricSnapshot{
		Username:            req.Username,
		GeneratedAt:         generatedAt,
		PositionCount:       len(positions),
		Financial:           financial,
		Source:              "read_model",
		TransactionSafeRead: false,
	}
}
