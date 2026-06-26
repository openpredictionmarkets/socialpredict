package markets

import (
	"context"
	"sort"
	"time"

	"socialpredict/internal/domain/boundary"
)

// GetMarketBets returns the bet history for a market with probabilities.
func (s *Service) GetMarketBets(ctx context.Context, marketID int64) ([]*BetDisplayInfo, error) {
	return s.getMarketBetDisplayInfos(ctx, marketID)
}

// GetMarketBetsPage returns a display page of market bets. Probability history
// is still derived from the full bet history so paginated display does not
// create a second probability math path.
func (s *Service) GetMarketBetsPage(ctx context.Context, marketID int64, p Page) ([]*BetDisplayInfo, error) {
	infos, err := s.getMarketBetDisplayInfos(ctx, marketID)
	if err != nil {
		return nil, err
	}
	sortBetDisplayInfosNewestFirst(infos)
	p = s.statusPolicy.NormalizePage(p, 20, 100)
	return paginateBetDisplayInfos(infos, p), nil
}

// ListBetsForMarket returns canonical market bet history for transaction-policy callers.
func (s *Service) ListBetsForMarket(ctx context.Context, marketID int64) ([]*Bet, error) {
	if marketID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.ListBetsForMarket(ctx, marketID)
}

func (s *Service) getMarketBetDisplayInfos(ctx context.Context, marketID int64) ([]*BetDisplayInfo, error) {
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

	modelBets, err := s.loadMarketBets(ctx, marketID)
	if err != nil {
		return nil, err
	}
	if len(modelBets) == 0 {
		return []*BetDisplayInfo{}, nil
	}

	probabilityChanges := ensureProbabilityChanges(s.probabilityEngine.Calculate(market.CreatedAt, modelBets), market.CreatedAt)
	sortProbabilityChanges(probabilityChanges)
	sortBetsByTime(modelBets)

	return buildBetDisplayInfos(modelBets, probabilityChanges), nil
}

func (s *Service) loadMarketBets(ctx context.Context, marketID int64) ([]boundary.Bet, error) {
	bets, err := s.repo.ListBetsForMarket(ctx, marketID)
	if err != nil {
		return nil, err
	}
	return convertToBoundaryBets(bets), nil
}

func ensureProbabilityChanges(changes []ProbabilityChange, createdAt time.Time) []ProbabilityChange {
	if len(changes) == 0 {
		return []ProbabilityChange{{
			Probability: 0,
			Timestamp:   createdAt,
		}}
	}
	return changes
}

func sortProbabilityChanges(changes []ProbabilityChange) {
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Timestamp.Before(changes[j].Timestamp)
	})
}

func sortBetsByTime(bets []boundary.Bet) {
	sort.Slice(bets, func(i, j int) bool {
		return bets[i].PlacedAt.Before(bets[j].PlacedAt)
	})
}

func buildBetDisplayInfos(boundaryBets []boundary.Bet, probabilityChanges []ProbabilityChange) []*BetDisplayInfo {
	results := make([]*BetDisplayInfo, 0, len(boundaryBets))
	for _, bet := range boundaryBets {
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

func sortBetDisplayInfosNewestFirst(infos []*BetDisplayInfo) {
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].PlacedAt.After(infos[j].PlacedAt)
	})
}

func paginateBetDisplayInfos(infos []*BetDisplayInfo, p Page) []*BetDisplayInfo {
	if len(infos) == 0 || p.Offset >= len(infos) {
		return []*BetDisplayInfo{}
	}
	end := p.Offset + p.Limit
	if end > len(infos) {
		end = len(infos)
	}
	return infos[p.Offset:end]
}

func latestProbabilityAt(changes []ProbabilityChange, timestamp time.Time) float64 {
	matched := changes[0].Probability
	for _, change := range changes {
		if change.Timestamp.After(timestamp) {
			break
		}
		matched = change.Probability
	}
	return matched
}
