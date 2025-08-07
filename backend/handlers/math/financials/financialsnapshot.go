package financials

import (
	positionsmath "socialpredict/handlers/math/positions"
	"socialpredict/setup"

	"gorm.io/gorm"
)

// ComputeUserFinancials calculates comprehensive financial metrics for a user
// using only existing models and stateless computations
func ComputeUserFinancials(db *gorm.DB, username string, accountBalance int64, econ *setup.EconomicConfig) (map[string]int64, error) {
	positions, err := positionsmath.CalculateAllUserMarketPositions_WPAM_DBPM(db, username)
	if err != nil {
		return nil, err
	}

	var (
		amountInPlay       int64 // Total current value across all positions
		amountInPlayActive int64 // Value in unresolved markets only
		totalSpent         int64 // Total amount ever spent
		totalSpentInPlay   int64 // Amount spent in unresolved markets
		tradingProfits     int64 // Total profits (realized + potential)
		realizedProfits    int64 // Profits from resolved markets
		potentialProfits   int64 // Profits from unresolved markets
		realizedValue      int64 // Final value from resolved positions
		potentialValue     int64 // Current value from unresolved positions
	)

	for _, pos := range positions {
		profit := pos.Value - pos.TotalSpent

		amountInPlay += pos.Value
		totalSpent += pos.TotalSpent
		tradingProfits += profit

		if pos.IsResolved {
			// Resolved market
			realizedProfits += profit
			realizedValue += pos.Value
		} else {
			// Unresolved market
			potentialProfits += profit
			potentialValue += pos.Value
			amountInPlayActive += pos.Value
			totalSpentInPlay += pos.TotalSpentInPlay
		}
	}

	workProfits, err := sumWorkProfitsFromTransactions(db, username)
	if err != nil {
		return nil, err
	}

	amountBorrowed := int64(0)
	if accountBalance < 0 {
		amountBorrowed = -accountBalance
	}

	retainedEarnings := accountBalance - amountInPlay
	equity := retainedEarnings + amountInPlay - amountBorrowed
	totalProfits := tradingProfits + workProfits

	return map[string]int64{
		// Original required fields from checkpoint
		"accountBalance":     accountBalance,
		"maximumDebtAllowed": econ.Economics.User.MaximumDebtAllowed,
		"amountInPlay":       amountInPlay,
		"amountBorrowed":     amountBorrowed,
		"retainedEarnings":   retainedEarnings,
		"equity":             equity,
		"tradingProfits":     tradingProfits,
		"workProfits":        workProfits,
		"totalProfits":       totalProfits,

		// Enhanced granular fields for potential vs realized breakdown
		"amountInPlayActive": amountInPlayActive, // Value in unresolved markets
		"totalSpent":         totalSpent,         // Total ever spent
		"totalSpentInPlay":   totalSpentInPlay,   // Spent in unresolved markets
		"realizedProfits":    realizedProfits,    // From resolved markets
		"potentialProfits":   potentialProfits,   // From unresolved markets
		"realizedValue":      realizedValue,      // Final payouts received
		"potentialValue":     potentialValue,     // Current unresolved value
	}, nil
}
