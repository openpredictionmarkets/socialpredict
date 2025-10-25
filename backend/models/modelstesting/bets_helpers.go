package modelstesting

import (
	"socialpredict/models"
	"socialpredict/setup"
)

type betKey struct {
	marketID uint
	username string
}

// CalculateParticipationFees determines the total participation fees owed for the provided
// bets according to the configured economics rules. Only the first positive bet per user
// per market incurs the fee.
func CalculateParticipationFees(cfg *setup.EconomicConfig, bets []models.Bet) int64 {
	var total int64
	seen := make(map[betKey]bool)
	initialFee := cfg.Economics.Betting.BetFees.InitialBetFee

	for _, bet := range bets {
		if bet.Amount <= 0 {
			continue
		}

		key := betKey{marketID: bet.MarketID, username: bet.Username}
		if !seen[key] {
			seen[key] = true
			total += initialFee
		}
	}

	return total
}
