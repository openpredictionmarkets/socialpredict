package marketmath

import "socialpredict/models"

// getMarketVolume returns the total volume of trades for a given market
func GetMarketVolume(bets []models.Bet) int64 {
	var totalVolume int64
	for _, bet := range bets {
		totalVolume += bet.Amount
	}

	totalVolumeUint := int64(totalVolume)

	return totalVolumeUint
}
