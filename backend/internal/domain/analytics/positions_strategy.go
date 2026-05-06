package analytics

import (
	"socialpredict/internal/domain/boundary"
	positionsmath "socialpredict/internal/domain/math/positions"
)

type defaultMarketPositionCalculator struct{}

func (defaultMarketPositionCalculator) Calculate(snapshot positionsmath.MarketSnapshot, bets []boundary.Bet) ([]positionsmath.MarketPosition, error) {
	return positionsmath.NewPositionCalculator().CalculateMarketPositions(snapshot, bets)
}

// NewMarketPositionCalculator builds a MarketPositionCalculator using the supplied math calculator.
func NewMarketPositionCalculator(calculator positionsmath.PositionCalculator) MarketPositionCalculator {
	return configurableMarketPositionCalculator{calculator: calculator}
}

type configurableMarketPositionCalculator struct {
	calculator positionsmath.PositionCalculator
}

func (c configurableMarketPositionCalculator) Calculate(snapshot positionsmath.MarketSnapshot, bets []boundary.Bet) ([]positionsmath.MarketPosition, error) {
	calc := c.calculator
	return calc.CalculateMarketPositions(snapshot, bets)
}
