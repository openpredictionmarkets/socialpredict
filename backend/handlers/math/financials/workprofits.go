package financials

import (
	"gorm.io/gorm"
)

// sumWorkProfitsFromTransactions calculates work-based profits from transactions
// Currently returns 0 since no separate transaction system exists yet (only bets)
// This maintains API structure for future extensibility when work rewards are added
func sumWorkProfitsFromTransactions(db *gorm.DB, username string) (int64, error) {
	// No separate transaction system exists yet - only models.Bet
	// Work rewards like "WorkReward" and "Bounty" types referenced in checkpoint don't exist
	// Return 0 to maintain financial snapshot structure for future extensibility
	return 0, nil
}
