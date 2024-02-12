package betshandlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	betutils "socialpredict/handlers/bets/betutils"
	marketshandlers "socialpredict/handlers/markets"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/handlers/tradingdata"
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

	// user-specific validation, sufficient balance,
	// Fetch the user's current balance
	if err := db.Where("username = ?", user.Username).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// sell opposite shares first
	// for betRequest of opposite Outcome held, first sell shares held at current price
	// Then go forward and adjust the betRequest to adjustedBetRequest and buy new shares with amount remaining
	// check users's current position on market, YES and NO
	marketIDStr := strconv.FormatUint(uint64(betRequest.MarketID), 10)
	quantityOppositeShares, oppositeDirection, err := marketshandlers.CheckOppositeSharesOwned(db, marketIDStr, user.Username, betRequest.Outcome)

	var betRequestRemain int64
	var totalPoolSaleQuanitity int64

	if quantityOppositeShares > 0 {
		// get the market information so we can extract the created at time to calculate price
		publicResponseMarket, err := marketshandlers.GetPublicResponseMarketByID(db, marketIDStr)
		if err != nil {
			http.Error(w, "Can't retrieve market", http.StatusBadRequest)
			return
		}

		// extract all bets on market to calculate current price
		allBetsOnMarket := tradingdata.GetBetsForMarket(db, betRequest.MarketID)
		allProbabilitiesOnMarket := wpam.CalculateMarketProbabilitiesWPAM(publicResponseMarket.CreatedAt, allBetsOnMarket)
		currentProbability := allProbabilitiesOnMarket[len(allProbabilitiesOnMarket)-1].Probability

		// Now, proceed to sell shares to the liquidity pool
		totalPoolSaleQuanitity, fee, err := SellSharesToPool(db, user.Username, quantityOppositeShares, oppositeDirection, currentProbability)
		if err != nil {
			// Handle error: Could not sell shares to the pool
			http.Error(w, "Failed to sell shares to the liquidity pool: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Adjust the total amount bet on the market after selling the shares, and fees
		betRequestRemain = betRequest.Amount - fee - totalPoolSaleQuanitity
	} else {
		betRequestRemain = betRequest.Amount
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

	// Deduct the bet and switching sides fee amount from the user's balance
	user.AccountBalance -= betRequestRemain

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

	// Save the Bet to the database
	result := db.Create(&bet)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Record the transaction for histocial data and for updating probability
	// Do this at the end, so that the pool bet further penalizes in the case that the person is switching sides
	// pool is buying the shares which were just sold
	if quantityOppositeShares > 0 {
		RecordPoolTransaction(db, betRequest.MarketID, betRequest.Outcome, totalPoolSaleQuanitity)
	}

	// Return a success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bet)
}

func SellSharesToPool(db *gorm.DB, username string, quantity int64, direction string, currentProbability float64) (int64, int64, error) {
	var pricePerShare float64
	if direction == "YES" {
		pricePerShare = currentProbability
	} else if direction == "NO" {
		pricePerShare = 1 - currentProbability
	}

	// Calculate the total sale value before fees
	totalSaleValueBeforeFees := float64(quantity) * pricePerShare

	var feePercent float64 = 0.05

	// Calculate the fee as 5% of the sale value
	fee := int64(math.Round(totalSaleValueBeforeFees * feePercent))

	// Ensure the fee is at least 1 point if 5% doesn't round up to 1
	if fee < 1 {
		fee = 1
	}

	// Adjust the total sale value after fees
	totalSaleValueAfterFees := totalSaleValueBeforeFees - float64(fee)

	// Round and convert to int64 for updating balance
	totalPoolSaleValue := int64(math.Round(totalSaleValueAfterFees))

	// Fetch the user from the database
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return totalPoolSaleValue, fee, err // User not found or other database error
	}

	// Update user's balance by adding the sale value, after fees
	user.AccountBalance += totalPoolSaleValue
	if err := db.Save(&user).Error; err != nil {
		return totalPoolSaleValue, fee, err // Error updating user balance
	}

	// Return the totalSaleValue for further processing
	return totalPoolSaleValue, fee, nil
}

// RecordPoolTransaction records a transaction made by the liquidity pool (admin) to the database.
func RecordPoolTransaction(db *gorm.DB, marketID uint, outcome string, totalSaleQuantity int64) error {
	// Assume "Admin" is the username for system-administered transactions.
	// You might have a dedicated admin account or system identifier for these operations.
	adminUsername := "admin"

	// Create a new Bet object with the admin as the user.
	// This bet reflects the liquidity pool's purchase or sale of shares.
	poolBet := models.Bet{
		Username: adminUsername, // Use a system identifier for the admin/user responsible for liquidity pool transactions.
		MarketID: marketID,
		Amount:   totalSaleQuantity, // The quantity of shares bought/sold by the pool.
		PlacedAt: time.Now(),        // Record the current time as the transaction time.
		Outcome:  outcome,           // The outcome for which the shares were bought/sold.
	}

	// Save the Bet to the database
	if result := db.Create(&poolBet); result.Error != nil {
		return fmt.Errorf("failed to record pool transaction: %w", result.Error)
	}

	return nil
}
