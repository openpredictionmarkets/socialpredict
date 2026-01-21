package analytics

import (
	positionsmath "socialpredict/internal/domain/math/positions"
	"socialpredict/models"
)

type defaultMarketPositionCalculator struct{}

func (defaultMarketPositionCalculator) Calculate(snapshot positionsmath.MarketSnapshot, bets []models.Bet) ([]positionsmath.MarketPosition, error) {
	return positionsmath.NewPositionCalculator().CalculateMarketPositions(snapshot, bets)
}

// NewMarketPositionCalculator builds a MarketPositionCalculator using the supplied math calculator.
func NewMarketPositionCalculator(calculator positionsmath.PositionCalculator) MarketPositionCalculator {
	return configurableMarketPositionCalculator{calculator: calculator}
}

type configurableMarketPositionCalculator struct {
	calculator positionsmath.PositionCalculator
}

func (c configurableMarketPositionCalculator) Calculate(snapshot positionsmath.MarketSnapshot, bets []models.Bet) ([]positionsmath.MarketPosition, error) {
	calc := c.calculator
	return calc.CalculateMarketPositions(snapshot, bets)
}
