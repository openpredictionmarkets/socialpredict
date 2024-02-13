package betshandlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	betutils "socialpredict/handlers/bets/betutils"
	marketshandlers "socialpredict/handlers/markets"
	"socialpredict/logging"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"
	"strconv"
	"time"

	"gorm.io/gorm"
)

func PlaceBetHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}

	// Validate JWT token and extract user information
	db := util.GetDB() // Get the database connection
	user, err := middleware.ValidateTokenAndGetUser(r, db)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var betRequest models.Bet
	// Decode the request body into betRequest
	err = json.NewDecoder(r.Body).Decode(&betRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the request (check if market exists, if not closed/resolved, etc.)
	betutils.CheckMarketStatus(db, betRequest.MarketID)

	// sell opposite shares first
	// for betRequest of opposite Outcome held, first sell shares held at current price
	// Then go forward and adjust the betRequest to adjustedBetRequest and buy new shares with amount remaining
	// check users's current position on market, YES and NO
	marketIDStr := strconv.FormatUint(uint64(betRequest.MarketID), 10)
	quantityOppositeShares, oppositeDirection, err := marketshandlers.CheckOppositeSharesOwned(db, marketIDStr, user.Username, betRequest.Outcome)

	logging.LogAnyType(quantityOppositeShares, "quantityOppositeShares")
	logging.LogAnyType(oppositeDirection, "oppositeDirection")

	var betRequestRemain int64
	var totalPoolSaleQuantity int64 = 0
	var fee int64
	var buySignal bool

	// if we have opposite shares, sell those first to pool, deduct fees, update user balance
	if quantityOppositeShares > 0 {

		// Now, proceed to sell shares to the liquidity pool, assess fee to user balance, but not add to trade
		totalPoolSaleQuantity, fee, buySignal, err = SellSharesToPool(db, betRequest, quantityOppositeShares, oppositeDirection)
		if err != nil {
			// Handle error: Could not sell shares to the pool
			http.Error(w, "Failed to sell shares to the liquidity pool: "+err.Error(), http.StatusInternalServerError)
			return
		}

		logging.LogAnyType(totalPoolSaleQuantity, "totalPoolSaleQuantity")
		logging.LogAnyType(fee, "fee")

		// Adjust the total amount bet on the market after selling the shares.
		// Fee is deducted from balance, not a part of the trade on the market.
		if buySignal {
			// if buySignal true denotes we still have left over to buy shares
			betRequestRemain = betRequest.Amount - totalPoolSaleQuantity
			logging.LogAnyType(betRequestRemain, "betRequestRemain")
		} else {
			// if buySignal false denotes we still have none leftover to buy additional shares
			betRequestRemain = 0
			logging.LogAnyType(betRequestRemain, "betRequestRemain")
		}

	} else {
		betRequestRemain = betRequest.Amount
		logging.LogAnyType(betRequestRemain, "betRequestRemain")
	}

	// Check if the user has enough balance to place the bet
	// Use the appConfig for configuration values
	maximumDebtAllowed := betutils.Appconfig.MaximumDebtAllowed

	// Check if the user's balance after the bet would be lower than the allowed maximum debt
	// deduct fee in case of switching sides
	if user.AccountBalance-betRequestRemain < -maximumDebtAllowed {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	logging.LogAnyType(user.AccountBalance, "user.AccountBalance before")
	// Deduct the bet and switching sides fee amount from the user's balance
	user.AccountBalance -= betRequestRemain
	logging.LogAnyType(user.AccountBalance, "user.AccountBalance after")

	// Update the user's balance in the database
	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Error updating user balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new Bet object
	bet := models.Bet{
		Username: user.Username,
		MarketID: betRequest.MarketID,
		Amount:   betRequestRemain,
		PlacedAt: time.Now(), // Set the current time as the placement time
		Outcome:  betRequest.Outcome,
	}

	logging.LogAnyType(bet, "bet")

	// Save the Bet to the database, if transaction was greater than 0.
	result := db.Create(&bet)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Record the transaction for histocial data and for updating probability
	// Do this at the end, so that the pool bet further penalizes in the case that the person is switching sides
	// pool is buying the shares which were just sold
	if quantityOppositeShares > 0 && totalPoolSaleQuantity > 0 {
		// quantityOppositeShares would make the AMM make a bet 100% in the opposite direction of whatever a human did, maintaining the price
		// but allowing the human to exit the market at that price, having to buy in the opposite direction to re-set it.
		// basically this makes it super onerous to switch positions, which might hypothetically disincentivize markets updating to be accurate
		logging.LogAnyType(quantityOppositeShares, "quantityOppositeShares in RecordPool if statement")
		// totalPoolSaleQuantity would make the AMM bet the difference between the shares owned and the new bet request, in the opposite direction
		// this is a way to only re-set the price down slightly on switch transactions, rather than all the way down to where it was previously upon exit.
		// basically this makes it more onerous to switch, slightly disincentivizing switching sides, but not completely.
		logging.LogAnyType(totalPoolSaleQuantity, "totalPoolSaleQuantity in RecordPool if statement")
		// here there is no disincentive to switch sides at all by the liquidity pool, the AMM does nothing and just assumes people can exit whenever.
		logging.LogAnyType(betRequestRemain, "betRequestRemain in RecordPool if statement")

		if betRequest.Amount < quantityOppositeShares {
			RecordPoolTransaction(db, betRequest.MarketID, oppositeDirection, betRequest.Amount)
		} else {
			RecordPoolTransaction(db, betRequest.MarketID, oppositeDirection, quantityOppositeShares)
		}
	}

	// Return a success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bet)
}

func SellSharesToPool(db *gorm.DB, betRequest models.Bet, quantityOppositeSharesHeld int64, direction string) (int64, int64, bool, error) {

	var differenceInRequestedAndHeld int64
	var feePercent float64 = 0.05
	var fee int64
	var saleDirection string

	// reverse the current buying direction to create a sale direction
	if betRequest.Outcome == "YES" {
		saleDirection = "NO"
	} else {
		saleDirection = "YES"
	}

	// Fetch the user from the database
	var user models.User
	if err := db.Where("username = ?", betRequest.Username).First(&user).Error; err != nil {
		return differenceInRequestedAndHeld, fee, false, err // User not found or other database error
	}

	differenceInRequestedAndHeld = int64(math.Abs(float64(betRequest.Amount - quantityOppositeSharesHeld)))
	fee = int64(math.Round(float64(differenceInRequestedAndHeld) * feePercent))
	// Ensure the fee is at least 1 point if 5% doesn't round up to 1
	if fee < 1 {
		fee = 1
	}

	// larger betRequest.Amount, need to sell opposite shares first and then buy leftover
	if betRequest.Amount > quantityOppositeSharesHeld {
		// Liquidate shares held by:
		// 1. Updating the user account balance, minus fee.
		// 2. Add transaction, quantity of quantityOppositeSharesHeld to bets as a sale (negative).
		// 3. Passing sale amount to differenceInRequestedAndHeld
		// 4. Passing boolean signal that we should continue forward with buying differenceInRequestedAndHeld
		user.AccountBalance += quantityOppositeSharesHeld - fee
		if err := db.Save(&user).Error; err != nil {
			return differenceInRequestedAndHeld, fee, false, err // Error updating user balance
		}

		// Create a new Bet object for selling, meaning negative of opposite shares held
		bet := models.Bet{
			Username: user.Username,
			MarketID: betRequest.MarketID,
			Amount:   -quantityOppositeSharesHeld,
			PlacedAt: time.Now(), // Set the current time as the placement time
			Outcome:  saleDirection,
		}

		logging.LogAnyType(bet, "bet")

		// Save the Bet to the database, if transaction was greater than 0.
		result := db.Create(&bet)
		if result.Error != nil {
			err := fmt.Errorf("failed to record opposite direction purchase transaction: %w", result.Error)
			return differenceInRequestedAndHeld, fee, false, err // Error updating user balance
		}

		// return differenceInRequestedAndHeld to be able to buy
		return differenceInRequestedAndHeld, fee, true, nil

	} else {
		// betRequest.Amount not sufficient to be greater than forced sales, so don't buy anything afterward.
		// Liquidate shares held by:
		// 1. Updating the user account balance, minus fee.
		// 2. Passing sale amount to differenceInRequestedAndHeld
		// 3. Passing boolean signal that we should NOT continue forward with buying differenceInRequestedAndHeld
		user.AccountBalance += betRequest.Amount - fee
		if err := db.Save(&user).Error; err != nil {
			return differenceInRequestedAndHeld, fee, false, err // Error updating user balance
		}

		// Create a new Bet object for selling, meaning negative of opposite shares held
		bet := models.Bet{
			Username: user.Username,
			MarketID: betRequest.MarketID,
			Amount:   -betRequest.Amount,
			PlacedAt: time.Now(), // Set the current time as the placement time
			Outcome:  saleDirection,
		}

		logging.LogAnyType(bet, "bet")

		// Save the Bet to the database, if transaction was greater than 0.
		result := db.Create(&bet)
		if result.Error != nil {
			err := fmt.Errorf("failed to record opposite direction purchase transaction: %w", result.Error)
			return differenceInRequestedAndHeld, fee, false, err // Error updating user balance
		}

		return differenceInRequestedAndHeld, fee, false, nil
	}

}

// RecordPoolTransaction records a transaction made by the liquidity pool (admin) to the database.
func RecordPoolTransaction(db *gorm.DB, marketID uint, outcome string, totalSaleQuantity int64) error {
	// Assume "Admin" is the username for system-administered transactions.
	// You might have a dedicated admin account or system identifier for these operations.
	poolUsername := "admin"

	// Create a new Bet object with the admin as the user.
	// This bet reflects the liquidity pool's purchase or sale of shares.
	poolBet := models.Bet{
		Username: poolUsername, // Use a system identifier for the admin/user responsible for liquidity pool transactions.
		MarketID: marketID,
		Amount:   totalSaleQuantity, // The quantity of shares bought/sold by the pool.
		PlacedAt: time.Now(),        // Record the current time as the transaction time.
		Outcome:  outcome,           // The outcome for which the shares were bought/sold.
	}

	logging.LogAnyType(poolBet.Amount, "RecordPoolTransaction poolBet.Amount")

	// Save the Bet to the database
	if result := db.Create(&poolBet); result.Error != nil {
		return fmt.Errorf("failed to record pool transaction: %w", result.Error)
	}

	return nil
}
