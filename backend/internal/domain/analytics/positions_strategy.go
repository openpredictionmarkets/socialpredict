package analytics

import (
	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"
)

type defaultMarketPositionCalculator struct{}

func (defaultMarketPositionCalculator) Calculate(snapshot positionsmath.MarketSnapshot, bets []models.Bet) ([]positionsmath.MarketPosition, error) {
	return positionsmath.NewPositionCalculator().CalculateMarketPositions(snapshot, bets)
}
