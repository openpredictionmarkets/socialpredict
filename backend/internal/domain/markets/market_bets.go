package markets

import (
	"context"
	"sort"
	"time"

	"socialpredict/internal/domain/math/probabilities/wpam"
	"socialpredict/models"
)

// GetMarketBets returns the bet history for a market with probabilities.
func (s *Service) GetMarketBets(ctx context.Context, marketID int64) ([]*BetDisplayInfo, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}

	market, err := s.repo.GetByID(ctx, marketID)
	if err != nil {
		return nil, err
	}

	modelBets, err := s.loadMarketBets(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if len(modelBets) == 0 {
		return []*BetDisplayInfo{}, nil
	}

	probabilityChanges := ensureProbabilityChanges(wpam.CalculateMarketProbabilitiesWPAM(market.CreatedAt, modelBets), market.CreatedAt)
	sortProbabilityChanges(probabilityChanges)
	sortBetsByTime(modelBets)

	return buildBetDisplayInfos(modelBets, probabilityChanges), nil
}

func (s *Service) loadMarketBets(ctx context.Context, marketID int64) ([]models.Bet, error) {
	bets, err := s.repo.ListBetsForMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}
	return convertToModelBets(bets), nil
}

func ensureProbabilityChanges(changes []wpam.ProbabilityChange, createdAt time.Time) []wpam.ProbabilityChange {
	if len(changes) == 0 {
		return []wpam.ProbabilityChange{{
			Probability: 0,
			Timestamp:   createdAt,
		}}
	}
	return changes
}

func sortProbabilityChanges(changes []wpam.ProbabilityChange) {
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Timestamp.Before(changes[j].Timestamp)
	})
}

func sortBetsByTime(bets []models.Bet) {
	sort.Slice(bets, func(i, j int) bool {
		return bets[i].PlacedAt.Before(bets[j].PlacedAt)
	})
}

func buildBetDisplayInfos(modelBets []models.Bet, probabilityChanges []wpam.ProbabilityChange) []*BetDisplayInfo {
	results := make([]*BetDisplayInfo, 0, len(modelBets))
	for _, bet := range modelBets {
		matchedProbability := latestProbabilityAt(probabilityChanges, bet.PlacedAt)
		results = append(results, &BetDisplayInfo{
			Username:    bet.Username,
			Outcome:     bet.Outcome,
			Amount:      bet.Amount,
			Probability: matchedProbability,
			PlacedAt:    bet.PlacedAt,
		})
	}
	return results
}

func latestProbabilityAt(changes []wpam.ProbabilityChange, timestamp time.Time) float64 {
	matched := changes[0].Probability
	for _, change := range changes {
		if change.Timestamp.After(timestamp) {
			break
		}
		matched = change.Probability
	}
	return matched
}
