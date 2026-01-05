package markets

import (
	"context"
)

// CalculateMarketVolume returns the total traded volume for a market.
func (s *Service) CalculateMarketVolume(ctx context.Context, marketID int64) (int64, error) {
	if marketID <= 0 {
		return 0, ErrInvalidInput
	}

	if _, err := s.repo.GetByID(ctx, marketID); err != nil {
		return 0, err
	}

	bets, err := s.repo.ListBetsForMarket(ctx, marketID)
	if err != nil {
		return 0, err
	}

	modelBets := convertToModelBets(bets)
	return s.metricsCalculator.Volume(modelBets), nil
}
