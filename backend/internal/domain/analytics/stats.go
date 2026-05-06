package analytics

import (
	"context"
	"errors"
)

// StatsConfig captures the economics values needed by stats reporting.
type StatsConfig struct {
	InitialAccountBalance int64
}

// FinancialStats captures financial aggregates exposed by stats reporting.
type FinancialStats struct {
	TotalMoney              int64
	TotalDebtExtended       int64
	TotalDebtUtilized       int64
	TotalFeesCollected      int64
	TotalBonusesPaid        int64
	OutstandingPayouts      int64
	TotalMoneyInCirculation int64
}

// ComputeFinancialStats returns the financial aggregates for the current system.
func (s *Service) ComputeFinancialStats(ctx context.Context, config StatsConfig) (FinancialStats, error) {
	if s == nil || s.statsRepo == nil {
		return FinancialStats{}, errors.New("stats repository not provided")
	}

	regularUsers, err := s.statsRepo.CountUsersByType(ctx, "REGULAR")
	if err != nil {
		return FinancialStats{}, err
	}

	return FinancialStats{
		TotalMoney: config.InitialAccountBalance * regularUsers,
	}, nil
}
