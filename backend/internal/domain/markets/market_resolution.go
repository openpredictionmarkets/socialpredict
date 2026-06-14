package markets

import (
	"context"

	users "socialpredict/internal/domain/users"
)

// ResolveMarket resolves a market with a given outcome.
// Resolution updates market state and user balances synchronously and remains
// outside background execution or retry infrastructure.
func (s *Service) ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error {
	return s.resolveMarket(ctx, marketID, resolution, username, true)
}

func (s *Service) resolveMarket(ctx context.Context, marketID int64, resolution string, username string, applyWorkProfit bool) error {
	outcome, err := s.resolutionPolicy.NormalizeResolution(resolution)
	if err != nil {
		return err
	}

	market, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return ErrMarketNotFound
	}
	if market == nil {
		return ErrMarketNotFound
	}

	if err := s.ensureMarketGovernanceActor(ctx, market, username); err != nil {
		return err
	}

	resolverUsername := username
	if !market.StewardedBy(username) {
		resolverUsername = market.CurrentStewardUsername()
	}
	if err := s.resolutionPolicy.ValidateResolutionRequest(market, resolverUsername); err != nil {
		return err
	}

	if err := s.resolutionPolicy.Resolve(ctx, s.repo, s.userService, marketID, outcome); err != nil {
		return err
	}

	if !applyWorkProfit {
		return nil
	}

	return s.applyModeratorWorkProfit(ctx, market, outcome, resolverUsername)
}

func (s *Service) applyModeratorWorkProfit(ctx context.Context, market *Market, outcome string, stewardUsername string) error {
	if market == nil || outcome == "N/A" || stewardUsername == "" || s.config.InitialBetFee <= 0 {
		return nil
	}

	income, err := s.calculateModeratorWorkProfitIncome(ctx, market)
	if err != nil {
		return err
	}
	if income <= 0 {
		return nil
	}

	return s.userService.ApplyTransaction(ctx, stewardUsername, income, users.TransactionWorkProfit)
}

func (s *Service) calculateModeratorWorkProfitIncome(ctx context.Context, market *Market) (int64, error) {
	if market == nil {
		return 0, nil
	}
	bets, err := s.repo.ListBetsForMarket(ctx, market.ID)
	if err != nil {
		return 0, err
	}
	return ModeratorWorkProfitIncome(bets, s.config.InitialBetFee, marketCreationCostForWorkProfit(market.ProposalCost, s.config.CreateMarketCost)), nil
}

// ModeratorWorkFeeIncome derives the first-participation fee income for a
// market from canonical bet history. Positive buy bets count once per unique
// participant; sell rows and later re-entry do not create additional income.
func ModeratorWorkFeeIncome(bets []*Bet, initialBetFee int64) int64 {
	if initialBetFee <= 0 {
		return 0
	}

	participants := make(map[string]struct{})
	for _, bet := range bets {
		if bet == nil || bet.Amount <= 0 || bet.Username == "" {
			continue
		}
		participants[bet.Username] = struct{}{}
	}

	return int64(len(participants)) * initialBetFee
}

// ModeratorWorkProfitIncome returns the work-profit payout after the market's
// creation cost threshold has been met. The first participation fees offset the
// market creation subsidy; only surplus fee income is paid to the steward.
func ModeratorWorkProfitIncome(bets []*Bet, initialBetFee int64, creationCost int64) int64 {
	income := ModeratorWorkFeeIncome(bets, initialBetFee) - creationCost
	if income < 0 {
		return 0
	}
	return income
}

func marketCreationCostForWorkProfit(proposalCost int64, fallbackCreateMarketCost int64) int64 {
	if proposalCost > 0 {
		return proposalCost
	}
	return fallbackCreateMarketCost
}
