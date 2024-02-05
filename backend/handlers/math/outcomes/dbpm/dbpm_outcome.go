package dbpm

import "socialpredict/models"

// calculatePayoutForOutcome calculates the payout for a specific outcome.
// betInput is the outcome of the bet (e.g., "YES", "NO").
// marketResolutionInput is the outcome to calculate the payout against (e.g., market resolution).
// See README/README-MATH-PROB-AND-PAYOUT.md#market-outcome-update-formulae---divergence-based-payout-model-dbpm
func CalculatePayoutForOutcomeDBPM(bet models.Bet, totalYes, totalNo int64, betInput, marketResolutionInput string) float64 {
	if betInput != marketResolutionInput {
		return 0 // No payout if the bet's outcome doesn't match the market resolution
	}

	var totalPoolForOutcome int64
	if marketResolutionInput == "YES" {
		totalPoolForOutcome = totalYes
	} else {
		totalPoolForOutcome = totalNo
	}

	totalPool := totalYes + totalNo

	return (bet.Amount / totalPoolForOutcome) * totalPool
}
