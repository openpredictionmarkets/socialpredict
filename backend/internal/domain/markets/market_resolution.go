package markets

import (
	"context"

	users "socialpredict/internal/domain/users"
)

// ResolveMarket resolves a market with a given outcome.
// Resolution updates market state and user balances synchronously and remains
// outside background execution or retry infrastructure.
func (s *Service) ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error {
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

	return s.applyModeratorWorkProfit(ctx, market, outcome, resolverUsername)
}

func (s *Service) applyModeratorWorkProfit(ctx context.Context, market *Market, outcome string, stewardUsername string) error {
	if market == nil || outcome == "N/A" || stewardUsername == "" || s.config.InitialBetFee <= 0 {
		return nil
	}

	income, err := s.calculateModeratorWorkFeeIncome(ctx, market.ID)
	if err != nil {
		return err
	}
	if income <= 0 {
		return nil
	}

	return s.userService.ApplyTransaction(ctx, stewardUsername, income, users.TransactionWorkProfit)
}

func (s *Service) calculateModeratorWorkFeeIncome(ctx context.Context, marketID int64) (int64, error) {
	bets, err := s.repo.ListBetsForMarket(ctx, marketID)
	if err != nil {
		return 0, err
	}
	return ModeratorWorkFeeIncome(bets, s.config.InitialBetFee), nil
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
