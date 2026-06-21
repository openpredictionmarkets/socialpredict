package analytics

import (
	"context"
	"errors"
	"time"

	"socialpredict/internal/domain/boundary"
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

	workProfits, err := s.computeUserWorkProfits(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	req.WorkProfits = workProfits
	unrealizedWorkIncome, unrealizedWorkProfits, err := s.computeUserUnrealizedWorkFinancials(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	req.UnrealizedWorkIncome = unrealizedWorkIncome
	req.UnrealizedWorkProfits = unrealizedWorkProfits

	snapshot := NewUserFinancialMetricSnapshotCalculator(s.config).Calculate(req, positions, time.Time{})
	return &snapshot.Financial, nil
}

func (s *Service) computeUserWorkProfits(ctx context.Context, username string) (int64, error) {
	if s.financialsRepo == nil {
		return 0, errors.New("financials repository not provided")
	}
	markets, err := s.financialsRepo.UserWorkProfitResolvedMarkets(ctx, username)
	if err != nil {
		return 0, err
	}

	var total int64
	for _, market := range markets {
		if !market.IsResolved || market.ResolutionResult == "N/A" {
			continue
		}
		bets, err := s.financialsRepo.ListBetsForMarket(ctx, market.ID)
		if err != nil {
			return 0, err
		}
		creationCost := creationCostForWorkProfit(market.ProposalCost, s.config.CreateMarketCost)
		total += stewardMarketWorkProfit(bets, s.config.InitialBetFee, creationCost)
	}

	groupRepo, ok := s.financialsRepo.(MarketGroupFinancialsRepository)
	if !ok {
		return total, nil
	}
	groups, err := groupRepo.UserWorkProfitResolvedMarketGroups(ctx, username)
	if err != nil {
		return 0, err
	}
	for _, group := range groups {
		if group.LifecycleStatus != "resolved" {
			continue
		}
		groupBets := make([][]boundary.Bet, 0, len(group.MemberMarketIDs))
		for _, marketID := range group.MemberMarketIDs {
			bets, err := s.financialsRepo.ListBetsForMarket(ctx, marketID)
			if err != nil {
				return 0, err
			}
			groupBets = append(groupBets, bets)
		}
		total += stewardMarketGroupWorkProfit(groupBets, s.config.InitialBetFee, creationCostForWorkProfit(group.ProposalCost, s.config.CreateMarketCost))
	}

	return total, nil
}

func (s *Service) computeUserUnrealizedWorkFinancials(ctx context.Context, username string) (int64, int64, error) {
	if s.financialsRepo == nil {
		return 0, 0, errors.New("financials repository not provided")
	}

	var income int64
	var profit int64
	marketRepo, ok := s.financialsRepo.(UnrealizedWorkProfitRepository)
	if ok {
		markets, err := marketRepo.UserWorkProfitUnresolvedMarkets(ctx, username)
		if err != nil {
			return 0, 0, err
		}
		for _, market := range markets {
			if market.IsResolved || market.ResolutionResult == "N/A" {
				continue
			}
			bets, err := s.financialsRepo.ListBetsForMarket(ctx, market.ID)
			if err != nil {
				return 0, 0, err
			}
			feeIncome := participationFeeIncome(bets, s.config.InitialBetFee)
			if market.StewardUsername == username {
				income += feeIncome
				profit += feeIncome
			}
			if market.CreatorUsername == username {
				profit -= creationCostForWorkProfit(market.ProposalCost, s.config.CreateMarketCost)
			}
		}
	}

	groupRepo, ok := s.financialsRepo.(UnrealizedMarketGroupFinancialsRepository)
	if !ok {
		return income, profit, nil
	}
	groups, err := groupRepo.UserWorkProfitUnresolvedMarketGroups(ctx, username)
	if err != nil {
		return 0, 0, err
	}
	for _, group := range groups {
		if group.LifecycleStatus == "resolved" || group.LifecycleStatus == "rejected" {
			continue
		}
		groupBets := make([][]boundary.Bet, 0, len(group.MemberMarketIDs))
		for _, marketID := range group.MemberMarketIDs {
			bets, err := s.financialsRepo.ListBetsForMarket(ctx, marketID)
			if err != nil {
				return 0, 0, err
			}
			groupBets = append(groupBets, bets)
		}
		feeIncome := groupParticipationFeeIncome(groupBets, s.config.InitialBetFee)
		if group.StewardUsername == username {
			income += feeIncome
			profit += feeIncome
		}
		if group.CreatorUsername == username {
			profit -= creationCostForWorkProfit(group.ProposalCost, s.config.CreateMarketCost)
		}
	}

	return income, profit, nil
}

func stewardMarketWorkProfit(bets []boundary.Bet, initialBetFee int64, creationCost int64) int64 {
	return participationFeeIncome(bets, initialBetFee) - creationCost
}

func participationFeeIncome(bets []boundary.Bet, initialBetFee int64) int64 {
	if initialBetFee <= 0 {
		return 0
	}
	participants := make(map[string]struct{})
	for _, bet := range bets {
		if bet.Amount <= 0 || bet.Username == "" {
			continue
		}
		participants[bet.Username] = struct{}{}
	}
	return int64(len(participants)) * initialBetFee
}

func stewardMarketGroupWorkProfit(betsByAnswer [][]boundary.Bet, initialBetFee int64, creationCost int64) int64 {
	return groupParticipationFeeIncome(betsByAnswer, initialBetFee) - creationCost
}

func groupParticipationFeeIncome(betsByAnswer [][]boundary.Bet, initialBetFee int64) int64 {
	if initialBetFee <= 0 {
		return 0
	}
	participants := make(map[string]struct{})
	for _, bets := range betsByAnswer {
		for _, bet := range bets {
			if bet.Amount <= 0 || bet.Username == "" {
				continue
			}
			participants[bet.Username] = struct{}{}
		}
	}
	return int64(len(participants)) * initialBetFee
}

func creationCostForWorkProfit(proposalCost int64, fallbackCreateMarketCost int64) int64 {
	if proposalCost > 0 {
		return proposalCost
	}
	return fallbackCreateMarketCost
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
		AccountBalance:        req.AccountBalance,
		MaximumDebtAllowed:    c.config.MaximumDebtAllowed,
		WorkProfits:           req.WorkProfits,
		UnrealizedWorkIncome:  req.UnrealizedWorkIncome,
		UnrealizedWorkProfits: req.UnrealizedWorkProfits,
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
