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

// returns the market volume + subsidization added into pool,
// subsidzation in pool could be paid out after resolution but not sold mid-market
func GetEndMarketVolume(bets []models.Bet, initialMarketSubsidization int64) int64 {
	return GetMarketVolume(bets) + initialMarketSubsidization
}
