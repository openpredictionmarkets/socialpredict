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

	groupRecords, err := listMarketGroupFeeRecords(ctx, repo)
	if err != nil {
		return nil, err
	}
	groupChildIDs := marketGroupChildIDSet(groupRecords)

	stats := &MarketVolumeStats{}
	for _, group := range groupRecords {
		stats.MarketCreationFees += creationCostForWorkProfit(group.ProposalCost, config.CreateMarketCost)
	}

	for _, market := range markets {
		if !groupChildIDs[market.ID] {
			stats.MarketCreationFees += creationCostForWorkProfit(market.ProposalCost, config.CreateMarketCost)
		}
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
	markets, err := repo.ListMarkets(ctx)
	if err != nil {
		return 0, err
	}
	betsOrdered, err := repo.ListBetsOrdered(ctx)
	if err != nil {
		return 0, err
	}
	groupRecords, err := listMarketGroupFeeRecords(ctx, repo)
	if err != nil {
		return 0, err
	}
	groupByID, childToGroup := marketGroupFeeLookups(groupRecords)

	marketByID := make(map[uint]MarketRecord, len(markets))
	for _, market := range markets {
		marketByID[market.ID] = market
	}

	feesByMarket := make(map[uint]int64)
	seen := make(map[struct {
		marketID uint
		username string
	}]bool)
	feesByGroup := make(map[uint]int64)
	seenGroup := make(map[struct {
		groupID  uint
		username string
	}]bool)

	for _, b := range betsOrdered {
		if b.Amount <= 0 {
			continue
		}
		if groupID, ok := childToGroup[b.MarketID]; ok {
			key := struct {
				groupID  uint
				username string
			}{groupID: groupID, username: b.Username}
			if !seenGroup[key] {
				feesByGroup[groupID] += config.InitialBetFee
				seenGroup[key] = true
			}
			continue
		}
		key := struct {
			marketID uint
			username string
		}{marketID: b.MarketID, username: b.Username}
		if !seen[key] {
			feesByMarket[b.MarketID] += config.InitialBetFee
			seen[key] = true
		}
	}

	var participationFees int64
	for marketID, feeIncome := range feesByMarket {
		market := marketByID[marketID]
		if market.IsResolved && market.ResolutionResult != "N/A" {
			participationFees += retainedParticipationFeesAfterWorkProfit(feeIncome, creationCostForWorkProfit(market.ProposalCost, config.CreateMarketCost))
			continue
		}
		participationFees += feeIncome
	}
	for groupID, feeIncome := range feesByGroup {
		group := groupByID[groupID]
		if group.LifecycleStatus == "resolved" {
			participationFees += retainedParticipationFeesAfterWorkProfit(feeIncome, creationCostForWorkProfit(group.ProposalCost, config.CreateMarketCost))
			continue
		}
		participationFees += feeIncome
	}

	return participationFees, nil
}

func listMarketGroupFeeRecords(ctx context.Context, repo any) ([]WorkProfitMarketGroupRecord, error) {
	groupRepo, ok := repo.(MarketGroupFeeRepository)
	if !ok {
		return nil, nil
	}
	return groupRepo.ListMarketGroupFeeRecords(ctx)
}

func marketGroupChildIDSet(groups []WorkProfitMarketGroupRecord) map[uint]bool {
	childIDs := make(map[uint]bool)
	for _, group := range groups {
		for _, marketID := range group.MemberMarketIDs {
			childIDs[marketID] = true
		}
	}
	return childIDs
}

func marketGroupFeeLookups(groups []WorkProfitMarketGroupRecord) (map[uint]WorkProfitMarketGroupRecord, map[uint]uint) {
	groupByID := make(map[uint]WorkProfitMarketGroupRecord, len(groups))
	childToGroup := make(map[uint]uint)
	for _, group := range groups {
		groupByID[group.ID] = group
		for _, marketID := range group.MemberMarketIDs {
			childToGroup[marketID] = group.ID
		}
	}
	return groupByID, childToGroup
}

func retainedParticipationFeesAfterWorkProfit(feeIncome int64, creationCost int64) int64 {
	if feeIncome <= 0 {
		return 0
	}
	if creationCost <= 0 {
		return 0
	}
	if feeIncome < creationCost {
		return feeIncome
	}
	return creationCost
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
			MarketCreationFees: NewInt64Metric(volume.MarketCreationFees, "Σ(standalone_market_proposal_costs) + Σ(group_proposal_costs)", "Fees collected from users creating standalone markets or grouped market proposals"),
			ParticipationFees:  NewInt64Metric(participationFees, "Σ(retained_first_bet_per_user_per_standalone_market_or_group × participation_fee)", "Retained fees collected from first-time participation; resolved markets and groups only redistribute surplus above the proposal-cost threshold as steward work profit"),
			BonusesPaid:        NewInt64Metric(bonusesPaid, "", "System bonuses paid to users and realized profits currently held in user balances"),
			TotalUtilized:      NewInt64Metric(totalUtilized, "unusedDebt + activeBetVolume + marketCreationFees + participationFees + bonusesPaid", "Total debt capacity that has been utilized across all categories"),
		},
		Verification: Verification{
			Balanced: NewBoolMetric(balanced, "Whether total created equals total utilized (perfect accounting balance)"),
			Surplus:  NewInt64Metric(surplus, "userDebtCapacity - totalUtilized", "Positive = unused capacity, Negative = over-utilization (indicates accounting error)"),
		},
	}
}
