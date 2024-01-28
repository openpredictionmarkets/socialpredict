package marketmath

import "socialpredict/models"

func CalculateTotalShares(bets []models.Bet, probabilityChanges []ProbabilityChange) uint {
	var totalShares uint = 0

	// Start from 1 since the first entry in probabilityChanges is just the initial condition
	for i, bet := range bets {
		// Since probabilityChanges has one extra entry at the beginning, use i+1
		probability := probabilityChanges[i+1].Probability

		if probability > 0 {
			shares := bet.Amount / probability
			totalShares += uint(shares)
		}
	}

	return totalShares
}
