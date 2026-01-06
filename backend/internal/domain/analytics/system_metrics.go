package analytics

import (
	"context"
	"errors"

	marketmath "socialpredict/internal/domain/math/market"
	"socialpredict/setup"
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
	if s.econLoader == nil {
		return nil, errors.New("economic configuration loader not provided")
	}

	s.ensureStrategyDefaults()
	econ := s.econLoader()

	debtStats, err := s.debtCalculator.Calculate(ctx, s.repo, econ)
	if err != nil {
		return nil, err
	}

	volumeStats, err := s.volumeCalculator.Calculate(ctx, s.repo, econ)
	if err != nil {
		return nil, err
	}

	participationFees, err := s.feeCalculator.CalculateParticipationFees(ctx, s.repo, econ)
	if err != nil {
		return nil, err
	}

	return s.metricsAssembler.Assemble(econ, debtStats, volumeStats, participationFees), nil
}

// DefaultDebtCalculator implements the existing debt policy.
type DefaultDebtCalculator struct{}

func (c DefaultDebtCalculator) Calculate(ctx context.Context, repo Repository, econ *setup.EconomicConfig) (*DebtStats, error) {
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
		stats.UnusedDebt += econ.Economics.User.MaximumDebtAllowed - usedDebt
	}

	stats.TotalDebtCapacity = econ.Economics.User.MaximumDebtAllowed * stats.UserCount
	return stats, nil
}

// DefaultVolumeCalculator implements the existing volume policy.
type DefaultVolumeCalculator struct{}

func (c DefaultVolumeCalculator) Calculate(ctx context.Context, repo Repository, econ *setup.EconomicConfig) (*MarketVolumeStats, error) {
	markets, err := repo.ListMarkets(ctx)
	if err != nil {
		return nil, err
	}

	stats := &MarketVolumeStats{
		MarketCreationFees: int64(len(markets)) * econ.Economics.MarketIncentives.CreateMarketCost,
	}

	for _, market := range markets {
		if market.IsResolved {
			continue
		}

		bets, err := repo.ListBetsForMarket(ctx, uint(market.ID))
		if err != nil {
			return nil, err
		}
		stats.ActiveBetVolume += marketmath.GetMarketVolume(bets)
	}

	return stats, nil
}

// DefaultFeeCalculator implements the existing participation fee policy.
type DefaultFeeCalculator struct{}

func (c DefaultFeeCalculator) CalculateParticipationFees(ctx context.Context, repo Repository, econ *setup.EconomicConfig) (int64, error) {
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
			participationFees += econ.Economics.Betting.BetFees.InitialBetFee
			seen[key] = true
		}
	}

	return participationFees, nil
}

// DefaultMetricsAssembler builds the SystemMetrics DTO from calculator outputs.
type DefaultMetricsAssembler struct{}

func (a DefaultMetricsAssembler) Assemble(econ *setup.EconomicConfig, debt *DebtStats, volume *MarketVolumeStats, participationFees int64) *SystemMetrics {
	bonusesPaid := debt.RealizedProfits
	totalUtilized := debt.UnusedDebt + volume.ActiveBetVolume + volume.MarketCreationFees + participationFees + bonusesPaid
	surplus := debt.TotalDebtCapacity - totalUtilized
	balanced := surplus == 0

	return &SystemMetrics{
		MoneyCreated: MoneyCreated{
			UserDebtCapacity: MetricWithExplanation{
				Value:       debt.TotalDebtCapacity,
				Formula:     "numUsers × maxDebtPerUser",
				Explanation: "Total credit capacity made available to all users",
			},
			NumUsers: MetricWithExplanation{
				Value:       debt.UserCount,
				Explanation: "Total number of registered users",
			},
		},
		MoneyUtilized: MoneyUtilized{
			UnusedDebt: MetricWithExplanation{
				Value:       debt.UnusedDebt,
				Formula:     "Σ(maxDebtPerUser - max(0, -balance))",
				Explanation: "Remaining borrowing capacity available to users",
			},
			ActiveBetVolume: MetricWithExplanation{
				Value:       volume.ActiveBetVolume,
				Formula:     "Σ(unresolved_market_volumes)",
				Explanation: "Total value of bets currently active in unresolved markets (excludes fees and subsidies)",
			},
			MarketCreationFees: MetricWithExplanation{
				Value:       volume.MarketCreationFees,
				Formula:     "number_of_markets × creation_fee_per_market",
				Explanation: "Fees collected from users creating new markets",
			},
			ParticipationFees: MetricWithExplanation{
				Value:       participationFees,
				Formula:     "Σ(first_bet_per_user_per_market × participation_fee)",
				Explanation: "Fees collected from first-time participation in each market",
			},
			BonusesPaid: MetricWithExplanation{
				Value:       bonusesPaid,
				Explanation: "System bonuses paid to users and realized profits currently held in user balances",
			},
			TotalUtilized: MetricWithExplanation{
				Value:       totalUtilized,
				Formula:     "unusedDebt + activeBetVolume + marketCreationFees + participationFees + bonusesPaid",
				Explanation: "Total debt capacity that has been utilized across all categories",
			},
		},
		Verification: Verification{
			Balanced: MetricWithExplanation{
				Value:       balanced,
				Explanation: "Whether total created equals total utilized (perfect accounting balance)",
			},
			Surplus: MetricWithExplanation{
				Value:       surplus,
				Formula:     "userDebtCapacity - totalUtilized",
				Explanation: "Positive = unused capacity, Negative = over-utilization (indicates accounting error)",
			},
		},
	}
}
