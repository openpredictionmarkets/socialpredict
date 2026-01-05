package markets

import (
	"context"
	"strings"

	"socialpredict/models"
)

// ProjectProbability projects what the probability would be after a hypothetical bet.
func (s *Service) ProjectProbability(ctx context.Context, req ProbabilityProjectionRequest) (*ProbabilityProjection, error) {
	if err := s.probabilityValidator.ValidateRequest(req); err != nil {
		return nil, err
	}

	market, err := s.repo.GetByID(ctx, req.MarketID)
	if err != nil {
		return nil, err
	}

	if err := s.probabilityValidator.ValidateMarket(market, s.clock.Now()); err != nil {
		return nil, err
	}

	bets, err := s.repo.ListBetsForMarket(ctx, req.MarketID)
	if err != nil {
		return nil, err
	}

	modelBets := convertToModelBets(bets)
	probabilityTrack := s.probabilityEngine.Calculate(market.CreatedAt, modelBets)

	currentProbability := 0.5
	if len(probabilityTrack) > 0 {
		currentProbability = probabilityTrack[len(probabilityTrack)-1].Probability
	}

	newBet := models.Bet{
		Username: "preview",
		MarketID: uint(market.ID),
		Amount:   req.Amount,
		Outcome:  normalizeOutcome(req.Outcome),
		PlacedAt: s.clock.Now(),
	}

	projection := s.probabilityEngine.Project(market.CreatedAt, modelBets, newBet)

	result := &ProbabilityProjection{
		CurrentProbability: currentProbability,
	}
	result.ProjectedProbability = projection.ProjectedProbability

	return result, nil
}

func normalizeOutcome(outcome string) string {
	switch strings.ToUpper(strings.TrimSpace(outcome)) {
	case "YES":
		return "YES"
	case "NO":
		return "NO"
	default:
		return ""
	}
}
