package marketmath

import "socialpredict/models"

// getMarketVolume returns the total volume of trades for a given market
func GetMarketVolume(bets []models.Bet) uint {
	var totalVolume float64
	for _, bet := range bets {
		totalVolume += bet.Amount
	}

	totalVolumeUint := uint(totalVolume)

	return totalVolumeUint
}
