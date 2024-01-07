package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/setup"
	"socialpredict/util"
	"time"

	"gorm.io/gorm"
)

// AppConfig holds the application-wide configuration
type AppConfig struct {
	InitialMarketProbability   float64
	InitialMarketSubsidization float64
	// user stuff
	MaximumDebtAllowed    float64
	InitialAccountBalance float64
	// betting stuff
	MinimumBet    float64
	BetFee        float64
	SellSharesFee float64
}

var appConfig AppConfig

func init() {
	// Load configuration
	config := setup.LoadEconomicsConfig()

	// Populate the appConfig struct
	appConfig = AppConfig{
		// market stuff
		InitialMarketProbability:   config.Economics.MarketCreation.InitialMarketProbability,
		InitialMarketSubsidization: config.Economics.MarketCreation.InitialMarketSubsidization,
		// user stuff
		MaximumDebtAllowed:    config.Economics.User.MaximumDebtAllowed,
		InitialAccountBalance: config.Economics.User.InitialAccountBalance,
		// betting stuff
		MinimumBet:    config.Economics.Betting.MinimumBet,
		BetFee:        config.Economics.Betting.BetFee,
		SellSharesFee: config.Economics.Betting.SellSharesFee,
	}
}

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

	// Validate the request (check if user and market exist, if amount is positive, etc.)
	// ...

	// Fetch the market to check if it is resolved
	var market models.Market
	if result := db.First(&market, betRequest.MarketID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Market not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching market", http.StatusInternalServerError)
		}
		return
	}

	// Check if the market is already resolved
	if market.IsResolved {
		http.Error(w, "Cannot place a bet on a resolved market", http.StatusBadRequest)
		return
	}

	// user-specific validation, sufficient balance,
	// Fetch the user's current balance
	if err := db.First(&user, user.ID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if the user has enough balance to place the bet

	// Use the appConfig for configuration values
	maximumDebtAllowed := appConfig.MaximumDebtAllowed

	// Check if the user's balance after the bet would be lower than the allowed maximum debt
	if user.AccountBalance-betRequest.Amount < -maximumDebtAllowed {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	// Deduct the bet amount from the user's balance
	user.AccountBalance -= betRequest.Amount

	// Update the user's balance in the database
	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Error updating user balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new Bet object
	bet := models.Bet{
		Username: user.Username,
		MarketID: betRequest.MarketID,
		Amount:   betRequest.Amount,
		PlacedAt: time.Now(), // Set the current time as the placement time
		Outcome:  betRequest.Outcome,
	}

	// Save the Bet to the database
	result := db.Create(&bet)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Return a success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bet)
}
