package markets

import (
	"context"

	"socialpredict/models"
)

// GetPublicMarket returns a public representation of a market.
func (s *Service) GetPublicMarket(ctx context.Context, marketID int64) (*PublicMarket, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.GetPublicMarket(ctx, marketID)
}

// ListMarkets returns a list of markets with filters.
func (s *Service) ListMarkets(ctx context.Context, filters ListFilters) ([]*Market, error) {
	return s.repo.List(ctx, filters)
}

// GetMarketOverviews returns enriched market data with calculations.
func (s *Service) GetMarketOverviews(ctx context.Context, filters ListFilters) ([]*MarketOverview, error) {
	markets, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, err
	}
	if markets == nil {
		return []*MarketOverview{}, nil
	}

	var overviews []*MarketOverview
	for _, market := range markets {
		overview := &MarketOverview{
			Market:  market,
			Creator: s.buildCreatorSummary(ctx, market.CreatorUsername),
		}
		overviews = append(overviews, overview)
	}

	return overviews, nil
}

// GetMarketDetails returns detailed market information with calculations.
func (s *Service) GetMarketDetails(ctx context.Context, marketID int64) (*MarketOverview, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}

	market, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if market == nil {
		return nil, ErrMarketNotFound
	}

	bets, err := s.repo.ListBetsForMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}

	modelBets := convertToModelBets(bets)
	probabilityChanges := s.probabilityEngine.Calculate(market.CreatedAt, modelBets)
	probabilityPoints := make([]ProbabilityPoint, len(probabilityChanges))
	for i, change := range probabilityChanges {
		probabilityPoints[i] = ProbabilityPoint{
			Probability: change.Probability,
			Timestamp:   change.Timestamp,
		}
	}

	lastProbability := 0.0
	if len(probabilityPoints) > 0 {
		lastProbability = probabilityPoints[len(probabilityPoints)-1].Probability
	}

	totalVolumeWithDust := s.metricsCalculator.VolumeWithDust(modelBets)
	marketDust := s.metricsCalculator.Dust(modelBets)
	numUsers := countUniqueUsers(modelBets)

	return &MarketOverview{
		Market:             market,
		Creator:            s.buildCreatorSummary(ctx, market.CreatorUsername),
		ProbabilityChanges: probabilityPoints,
		LastProbability:    lastProbability,
		NumUsers:           numUsers,
		TotalVolume:        totalVolumeWithDust,
		MarketDust:         marketDust,
	}, nil
}

func (s *Service) buildCreatorSummary(ctx context.Context, username string) *CreatorSummary {
	summary := &CreatorSummary{Username: username}
	if s.userService == nil {
		return summary
	}
	user, err := s.userService.GetPublicUser(ctx, username)
	if err != nil || user == nil {
		return summary
	}
	summary.DisplayName = user.DisplayName
	summary.PersonalEmoji = user.PersonalEmoji
	return summary
}

func convertToModelBets(bets []*Bet) []models.Bet {
	if len(bets) == 0 {
		return []models.Bet{}
	}
	out := make([]models.Bet, len(bets))
	for i, bet := range bets {
		out[i] = models.Bet{
			Username: bet.Username,
			MarketID: bet.MarketID,
			Amount:   bet.Amount,
			PlacedAt: bet.PlacedAt,
			Outcome:  bet.Outcome,
		}
	}
	return out
}

func countUniqueUsers(bets []models.Bet) int {
	if len(bets) == 0 {
		return 0
	}
	seen := make(map[string]struct{})
	for _, bet := range bets {
		if bet.Username == "" {
			continue
		}
		if _, ok := seen[bet.Username]; !ok {
			seen[bet.Username] = struct{}{}
		}
	}
	return len(seen)
}
