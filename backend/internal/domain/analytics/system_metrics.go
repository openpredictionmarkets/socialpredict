package analytics

import (
	"context"
	"errors"

	marketmath "socialpredict/internal/domain/math/market"
)

// DebtStats represents aggregated debt metrics.
type DebtStats struct {
	UserCount         int64
	UnusedDebt        int64
	RealizedProfits   int64
	TotalDebtCapacity int64
}

// MarketVolumeStats represents market volume metrics.
type MarketVolumeStats struct {
	MarketCreationFees int64
	ActiveBetVolume    int64
}

// ComputeSystemMetrics aggregates system-wide monetary metrics.
func (s *Service) ComputeSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
	if s.repo == nil {
		return nil, errors.New("repository not provided")
	}

	s.ensureStrategyDefaults()

	debtStats, err := s.debtCalculator.Calculate(ctx, s.repo, s.config)
	if err != nil {
		return nil, err
	}

	volumeStats, err := s.volumeCalculator.Calculate(ctx, s.repo, s.config)
	if err != nil {
		return nil, err
	}

	participationFees, err := s.feeCalculator.CalculateParticipationFees(ctx, s.repo, s.config)
	if err != nil {
		return nil, err
	}

	return s.metricsAssembler.Assemble(debtStats, volumeStats, participationFees), nil
}

// DefaultDebtCalculator implements the existing debt policy.
type DefaultDebtCalculator struct{}

func (c DefaultDebtCalculator) Calculate(ctx context.Context, repo DebtRepository, config Config) (*DebtStats, error) {
	users, err := repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	stats := &DebtStats{
		UserCount: int64(len(users)),
	}

	for _, user := range users {
		balance := user.AccountBalance
		if balance > 0 {
			stats.RealizedProfits += balance
		}
		usedDebt := int64(0)
		if balance < 0 {
			usedDebt = -balance
		}
		stats.UnusedDebt += config.MaximumDebtAllowed - usedDebt
	}

	stats.TotalDebtCapacity = config.MaximumDebtAllowed * stats.UserCount
	return stats, nil
}

// DefaultVolumeCalculator implements the existing volume policy.
type DefaultVolumeCalculator struct{}

func (c DefaultVolumeCalculator) Calculate(ctx context.Context, repo VolumeRepository, config Config) (*MarketVolumeStats, error) {
	markets, err := repo.ListMarkets(ctx)
	if err != nil {
		return nil, err
	}

	stats := &MarketVolumeStats{
		MarketCreationFees: int64(len(markets)) * config.CreateMarketCost,
	}

	for _, market := range markets {
		if market.IsResolved {
			continue
		}

		bets, err := repo.ListBetsForMarket(ctx, market.ID)
		if err != nil {
			return nil, err
		}
		stats.ActiveBetVolume += marketmath.GetMarketVolume(bets)
	}

	return stats, nil
}

// DefaultFeeCalculator implements the existing participation fee policy.
type DefaultFeeCalculator struct{}

func (c DefaultFeeCalculator) CalculateParticipationFees(ctx context.Context, repo FeeRepository, config Config) (int64, error) {
	betsOrdered, err := repo.ListBetsOrdered(ctx)
	if err != nil {
		return 0, err
	}

	type userMarket struct {
		marketID uint
		username string
	}

	seen := make(map[userMarket]bool)
	var participationFees int64

	for _, b := range betsOrdered {
		if b.Amount <= 0 {
			continue
		}
		key := userMarket{marketID: b.MarketID, username: b.Username}
		if !seen[key] {
			participationFees += config.InitialBetFee
			seen[key] = true
		}
	}

	return participationFees, nil
}

// DefaultMetricsAssembler builds the SystemMetrics DTO from calculator outputs.
type DefaultMetricsAssembler struct{}

func (a DefaultMetricsAssembler) Assemble(debt *DebtStats, volume *MarketVolumeStats, participationFees int64) *SystemMetrics {
	bonusesPaid := debt.RealizedProfits
	totalUtilized := debt.UnusedDebt + volume.ActiveBetVolume + volume.MarketCreationFees + participationFees + bonusesPaid
	surplus := debt.TotalDebtCapacity - totalUtilized
	balanced := surplus == 0

	return &SystemMetrics{
		MoneyCreated: MoneyCreated{
			UserDebtCapacity: NewInt64Metric(debt.TotalDebtCapacity, "numUsers × maxDebtPerUser", "Total credit capacity made available to all users"),
			NumUsers:         NewInt64Metric(debt.UserCount, "", "Total number of registered users"),
		},
		MoneyUtilized: MoneyUtilized{
			UnusedDebt:         NewInt64Metric(debt.UnusedDebt, "Σ(maxDebtPerUser - max(0, -balance))", "Remaining borrowing capacity available to users"),
			ActiveBetVolume:    NewInt64Metric(volume.ActiveBetVolume, "Σ(unresolved_market_volumes)", "Total value of bets currently active in unresolved markets (excludes fees and subsidies)"),
			MarketCreationFees: NewInt64Metric(volume.MarketCreationFees, "number_of_markets × creation_fee_per_market", "Fees collected from users creating new markets"),
			ParticipationFees:  NewInt64Metric(participationFees, "Σ(first_bet_per_user_per_market × participation_fee)", "Fees collected from first-time participation in each market"),
			BonusesPaid:        NewInt64Metric(bonusesPaid, "", "System bonuses paid to users and realized profits currently held in user balances"),
			TotalUtilized:      NewInt64Metric(totalUtilized, "unusedDebt + activeBetVolume + marketCreationFees + participationFees + bonusesPaid", "Total debt capacity that has been utilized across all categories"),
		},
		Verification: Verification{
			Balanced: NewBoolMetric(balanced, "Whether total created equals total utilized (perfect accounting balance)"),
			Surplus:  NewInt64Metric(surplus, "userDebtCapacity - totalUtilized", "Positive = unused capacity, Negative = over-utilization (indicates accounting error)"),
		},
	}
}
