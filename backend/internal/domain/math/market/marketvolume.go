package marketmath

import "socialpredict/models"

// VolumeCalculator abstracts market volume aggregation so alternate volume models can be substituted.
type VolumeCalculator interface {
	Volume(bets []models.Bet) int64
}

// EndVolumeCalculator abstracts final pool volume calculations that include subsidization.
type EndVolumeCalculator interface {
	EndVolume(bets []models.Bet, initialMarketSubsidization int64) int64
}

type summingVolumeCalculator struct{}

var defaultVolumeCalculator VolumeCalculator = summingVolumeCalculator{}
var defaultEndVolumeCalculator EndVolumeCalculator = summingVolumeCalculator{}

// GetMarketVolume returns the total volume of trades for a given market.
func GetMarketVolume(bets []models.Bet) int64 {
	return defaultVolumeCalculator.Volume(bets)
}

// GetEndMarketVolume returns market volume plus subsidization added into the pool.
func GetEndMarketVolume(bets []models.Bet, initialMarketSubsidization int64) int64 {
	return defaultEndVolumeCalculator.EndVolume(bets, initialMarketSubsidization)
}

func (summingVolumeCalculator) Volume(bets []models.Bet) int64 {
	var totalVolume int64
	for _, bet := range bets {
		totalVolume += bet.Amount
	}
	return totalVolume
}

func (calculator summingVolumeCalculator) EndVolume(bets []models.Bet, initialMarketSubsidization int64) int64 {
	return calculator.Volume(bets) + initialMarketSubsidization
}
