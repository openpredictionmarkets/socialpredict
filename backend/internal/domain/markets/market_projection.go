package markets

import (
	"context"
	"math"
	"strings"
	"time"

	"socialpredict/internal/domain/math/probabilities/wpam"
	"socialpredict/models"
)

// ProjectProbability projects what the probability would be after a hypothetical bet.
func (s *Service) ProjectProbability(ctx context.Context, req ProbabilityProjectionRequest) (*ProbabilityProjection, error) {
	if err := validateProbabilityRequest(req); err != nil {
		return nil, err
	}

	market, err := s.repo.GetByID(ctx, req.MarketID)
	if err != nil {
		return nil, err
	}

	if err := validateMarketForProjection(market, s.clock.Now()); err != nil {
		return nil, err
	}

	bets, err := s.repo.ListBetsForMarket(ctx, req.MarketID)
	if err != nil {
		return nil, err
	}

	projectionInput := projectionInputs{
		market:    market,
		outcome:   normalizeOutcome(req.Outcome),
		amount:    req.Amount,
		now:       s.clock.Now(),
		modelBets: convertToModelBets(bets),
	}

	return calculateProbabilityProjection(projectionInput), nil
}

type projectionInputs struct {
	market    *Market
	outcome   string
	amount    int64
	now       time.Time
	modelBets []models.Bet
}

func validateProbabilityRequest(req ProbabilityProjectionRequest) error {
	if req.MarketID <= 0 || req.MarketID > int64(math.MaxUint32) || strings.TrimSpace(req.Outcome) == "" || req.Amount <= 0 {
		return ErrInvalidInput
	}

	outcome := strings.ToUpper(strings.TrimSpace(req.Outcome))
	if outcome != "YES" && outcome != "NO" {
		return ErrInvalidInput
	}
	return nil
}

func validateMarketForProjection(market *Market, now time.Time) error {
	if strings.EqualFold(market.Status, "resolved") {
		return ErrInvalidState
	}

	if now.After(market.ResolutionDateTime) {
		return ErrInvalidState
	}
	return nil
}

func calculateProbabilityProjection(input projectionInputs) *ProbabilityProjection {
	probabilityTrack := wpam.CalculateMarketProbabilitiesWPAM(input.market.CreatedAt, input.modelBets)

	currentProbability := 0.5
	if len(probabilityTrack) > 0 {
		currentProbability = probabilityTrack[len(probabilityTrack)-1].Probability
	}

	newBet := models.Bet{
		Username: "preview",
		MarketID: uint(input.market.ID),
		Amount:   input.amount,
		Outcome:  input.outcome,
		PlacedAt: input.now,
	}

	projection := wpam.ProjectNewProbabilityWPAM(input.market.CreatedAt, input.modelBets, newBet)

	return &ProbabilityProjection{
		CurrentProbability:   currentProbability,
		ProjectedProbability: projection.Probability,
	}
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
