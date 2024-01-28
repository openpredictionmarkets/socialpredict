package marketmath

import "socialpredict/models"

// getMarketVolume returns the total volume of trades for a given market
func GetMarketVolume(bets []models.Bet) float64 {
	var totalVolume float64
	for _, bet := range bets {
		totalVolume += bet.Amount
	}

	return totalVolume
}
