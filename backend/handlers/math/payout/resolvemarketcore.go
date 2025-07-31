package payout

import (
	"errors"
	"fmt"
	positionsmath "socialpredict/handlers/math/positions"
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
	case "PROB":
		return fmt.Errorf("probabilistic resolution is not yet supported")
	default:
		return fmt.Errorf("unsupported resolution result: %q", market.ResolutionResult)
	}
}

func calculateAndAllocateProportionalPayouts(market *models.Market, db *gorm.DB) error {
	// Step 1: Convert market ID formats
	marketIDStr := strconv.FormatInt(market.ID, 10)

	// Step 2: Calculate market positions with resolved valuation
	displayPositions, err := positionsmath.CalculateMarketPositions_WPAM_DBPM(db, marketIDStr)
	if err != nil {
		return err
	}

	// Step 3: Pay out each user their resolved valuation
	for _, pos := range displayPositions {
		if pos.Value > 0 {
			if err := usersHandlers.ApplyTransactionToUser(pos.Username, pos.Value, db, usersHandlers.TransactionWin); err != nil {
				return err
			}
		}
	}

	return nil
}

func refundAllBets(market *models.Market, db *gorm.DB) error {
	// Retrieve all bets associated with the market
	var bets []models.Bet
	if err := db.Where("market_id = ?", market.ID).Find(&bets).Error; err != nil {
		return err
	}

	// Refund each bet to the user
	for _, bet := range bets {
		if err := usersHandlers.ApplyTransactionToUser(bet.Username, bet.Amount, db, usersHandlers.TransactionRefund); err != nil {
			return err
		}
	}

	return nil
}
