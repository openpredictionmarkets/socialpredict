package payout

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	positionsmath "socialpredict/handlers/math/positions"
	dusers "socialpredict/internal/domain/users"
	rusers "socialpredict/internal/repository/users"
	"socialpredict/models"

	"gorm.io/gorm"
)

func DistributePayoutsWithRefund(market *models.Market, db *gorm.DB) error {
	if market == nil {
		return errors.New("market is nil")
	}

	usersService := dusers.NewService(rusers.NewGormRepository(db), nil, nil)

	switch market.ResolutionResult {
	case "N/A":
		return refundAllBets(context.Background(), market, db, usersService)
	case "YES", "NO":
		return calculateAndAllocateProportionalPayouts(context.Background(), market, db, usersService)
	case "PROB":
		return fmt.Errorf("probabilistic resolution is not yet supported")
	default:
		return fmt.Errorf("unsupported resolution result: %q", market.ResolutionResult)
	}
}

func calculateAndAllocateProportionalPayouts(ctx context.Context, market *models.Market, db *gorm.DB, usersService dusers.ServiceInterface) error {
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
			if err := usersService.ApplyTransaction(ctx, pos.Username, pos.Value, dusers.TransactionWin); err != nil {
				return err
			}
		}
	}

	return nil
}

func refundAllBets(ctx context.Context, market *models.Market, db *gorm.DB, usersService dusers.ServiceInterface) error {
	// Retrieve all bets associated with the market
	var bets []models.Bet
	if err := db.Where("market_id = ?", market.ID).Find(&bets).Error; err != nil {
		return err
	}

	// Refund each bet to the user
	for _, bet := range bets {
		if err := usersService.ApplyTransaction(ctx, bet.Username, bet.Amount, dusers.TransactionRefund); err != nil {
			return err
		}
	}

	return nil
}
