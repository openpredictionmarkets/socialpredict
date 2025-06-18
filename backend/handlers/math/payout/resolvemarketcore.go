package payout

import (
	"errors"
	"fmt"
	marketmath "socialpredict/handlers/math/market"
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/handlers/positions"
	usersHandlers "socialpredict/handlers/users"
	"socialpredict/models"
	"strconv"

	"gorm.io/gorm"
)

func DistributePayoutsWithRefund(market *models.Market, db *gorm.DB) error {
	if market == nil {
		return errors.New("market is nil")
	}

	switch market.ResolutionResult {
	case "N/A":
		return refundAllBets(market, db)
	case "YES", "NO":
		return calculateAndAllocateProportionalPayouts(market, db)
	default:
		return fmt.Errorf("unsupported resolution result: %s", market.ResolutionResult)
	}
}

func refundAllBets(market *models.Market, db *gorm.DB) error {
	var bets []models.Bet
	if err := db.Where("market_id = ?", market.ID).Find(&bets).Error; err != nil {
		return err
	}
	for _, bet := range bets {
		if err := usersHandlers.UpdateUserBalance(bet.Username, bet.Amount, db, "refund"); err != nil {
			return err
		}
	}
	return nil
}

func calculateAndAllocateProportionalPayouts(market *models.Market, db *gorm.DB) error {
	bets := []models.Bet{}
	if err := db.Where("market_id = ?", market.ID).Find(&bets).Error; err != nil {
		return err
	}

	totalVolume := marketmath.GetMarketVolume(bets)

	positionsRaw, err := positions.CalculateMarketPositions_WPAM_DBPM(db, strconv.FormatInt(market.ID, 10))
	if err != nil {
		return err
	}

	var dbpmPositions []dbpm.DBPMMarketPosition
	for _, p := range positionsRaw {
		dbpmPositions = append(dbpmPositions, dbpm.DBPMMarketPosition{
			Username:       p.Username,
			NoSharesOwned:  p.NoSharesOwned,
			YesSharesOwned: p.YesSharesOwned,
			// Don't include Value field; it's not part of DBPMMarketPosition
		})
	}

	winningPositions, totalWinningShares := SelectWinningPositions(market.ResolutionResult, dbpmPositions)

	if totalWinningShares == 0 {
		return nil
	}

	AllocateWinningSharePool(db, market, winningPositions, totalWinningShares, totalVolume)

	return nil
}
