package marketshandlers

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/handlers/positions"
	usersHandlers "socialpredict/handlers/users"
	"socialpredict/models"

	"gorm.io/gorm"
)

func SelectWinningPositions(resolution string, raw []dbpm.DBPMMarketPosition) ([]positions.MarketPosition, int64) {
	var winningPositions []positions.MarketPosition
	var totalWinningShares int64

	for _, pos := range raw {
		converted := positions.MarketPosition{
			Username:       pos.Username,
			YesSharesOwned: pos.YesSharesOwned,
			NoSharesOwned:  pos.NoSharesOwned,
		}

		if resolution == "YES" && pos.YesSharesOwned > 0 {
			totalWinningShares += pos.YesSharesOwned
			winningPositions = append(winningPositions, converted)
		} else if resolution == "NO" && pos.NoSharesOwned > 0 {
			totalWinningShares += pos.NoSharesOwned
			winningPositions = append(winningPositions, converted)
		}
	}

	return winningPositions, totalWinningShares
}

func AllocateWinningSharePool(
	db *gorm.DB,
	market *models.Market,
	winningPositions []positions.MarketPosition,
	totalWinningShares int64,
	totalVolume int64,
) error {
	type payoutResult struct {
		username string
		payout   int64
		shares   int64
	}

	var (
		results      []payoutResult
		sumPayouts   int64
		maxShares    int64
		topUserIndex int
	)

	// Calculate payouts and track user with largest share
	for i, pos := range winningPositions {
		var userShares int64
		if market.ResolutionResult == "YES" {
			userShares = pos.YesSharesOwned
		} else {
			userShares = pos.NoSharesOwned
		}

		payout := int64(float64(userShares) / float64(totalWinningShares) * float64(totalVolume))
		sumPayouts += payout

		if userShares > maxShares {
			maxShares = userShares
			topUserIndex = i
		}

		results = append(results, payoutResult{
			username: pos.Username,
			payout:   payout,
			shares:   userShares,
		})
	}

	// Calculate delta from rounding and assign remainder to top user
	delta := totalVolume - sumPayouts
	if delta > 0 {
		results[topUserIndex].payout += delta
	}

	// Apply payouts to user balances
	for _, res := range results {
		if res.payout > 0 {
			if err := usersHandlers.UpdateUserBalance(res.username, res.payout, db, "win"); err != nil {
				return err
			}
		}
	}

	return nil
}
