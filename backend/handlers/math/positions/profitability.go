package positionsmath

import (
	"errors"
	"log"
	"math"
	"socialpredict/handlers/tradingdata"
	"socialpredict/models"
	"sort"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// Define a constant for the maximum value of uint for static analysis (CodeQL)
const maxUintValue uint64 = 4294967295 // For 32-bit systems; adjust for 64-bit if needed
// UserProfitability represents a user's profitability data for a specific market
type UserProfitability struct {
	Username       string    `json:"username"`
	CurrentValue   int64     `json:"currentValue"`
	TotalSpent     int64     `json:"totalSpent"`
	Profit         int64     `json:"profit"`
	Position       string    `json:"position"` // "YES", "NO", "NEUTRAL"
	YesSharesOwned int64     `json:"yesSharesOwned"`
	NoSharesOwned  int64     `json:"noSharesOwned"`
	EarliestBet    time.Time `json:"earliestBet"`
	Rank           int       `json:"rank"`
}

// ErrorLogger logs an error and returns a boolean indicating whether an error occurred.
func ErrorLogger(err error, errMsg string) bool {
	if err != nil {
		log.Printf("Error: %s - %s\n", errMsg, err) // Combine your custom message with the error's message.
		return true                                 // Indicate that an error was handled.
	}
	return false // No error to handle.
}

// CalculateUserSpend calculates the total amount a user has spent on a market
// by summing all positive amounts (purchases) and subtracting negative amounts (sales)
func CalculateUserSpend(bets []models.Bet, username string) int64 {
	var totalSpend int64 = 0

	for _, bet := range bets {
		if bet.Username == username {
			totalSpend += bet.Amount // Amount can be positive (buy) or negative (sell)
		}
	}

	return totalSpend
}

// GetEarliestBetTime finds the earliest bet timestamp for a user in a market
// Used as a tiebreaker for ranking users with identical profitability
func GetEarliestBetTime(bets []models.Bet, username string) time.Time {
	var earliestTime time.Time
	found := false

	for _, bet := range bets {
		if bet.Username == username {
			if !found || bet.PlacedAt.Before(earliestTime) {
				earliestTime = bet.PlacedAt
				found = true
			}
		}
	}

	return earliestTime
}

// DeterminePositionType determines if a user is holding YES, NO, or NEUTRAL positions
func DeterminePositionType(yesShares, noShares int64) string {
	if yesShares > 0 && noShares == 0 {
		return "YES"
	} else if noShares > 0 && yesShares == 0 {
		return "NO"
	} else if yesShares > 0 && noShares > 0 {
		return "NEUTRAL"
	}
	// This case shouldn't happen since we filter out zero positions
	return "NONE"
}

// CalculateMarketLeaderboard calculates profitability rankings for all users with positions in a market
func CalculateMarketLeaderboard(db *gorm.DB, marketIdStr string) ([]UserProfitability, error) {
	// Convert marketId string to uint64
	marketIDUint64, err := strconv.ParseUint(marketIdStr, 10, 64)
	if err != nil {
		ErrorLogger(err, "Can't convert marketIdStr to uint64.")
		return nil, err
	}

	// Check that marketIDUint64 fits in uint using explicit constant bound (security vulnerability fix)
	if marketIDUint64 > maxUintValue {
		err := errors.New("marketId out of range for uint")
		ErrorLogger(err, "marketIdStr is too large for uint.")
		return nil, err
	}

	marketIDUint := uint(marketIDUint64)

	// Get current positions and values using existing function
	marketPositions, err := CalculateMarketPositions_WPAM_DBPM(db, marketIdStr)
	if err != nil {
		ErrorLogger(err, "Failed to calculate market positions.")
		return nil, err
	}

	// Get all bets for the market to calculate spend
	allBetsOnMarket := tradingdata.GetBetsForMarket(db, marketIDUint)
	if len(allBetsOnMarket) == 0 {
		return []UserProfitability{}, nil
	}

	// Calculate profitability for each user with positions
	var leaderboard []UserProfitability

	for _, position := range marketPositions {
		// Filter out users with zero positions (no current stake in market)
		if position.YesSharesOwned == 0 && position.NoSharesOwned == 0 {
			continue
		}

		// Calculate total spend for this user
		totalSpent := CalculateUserSpend(allBetsOnMarket, position.Username)

		// Calculate profit = current value - total spent
		profit := position.Value - totalSpent

		// Determine position type
		positionType := DeterminePositionType(position.YesSharesOwned, position.NoSharesOwned)

		// Get earliest bet time for tiebreaker
		earliestBet := GetEarliestBetTime(allBetsOnMarket, position.Username)

		leaderboard = append(leaderboard, UserProfitability{
			Username:       position.Username,
			CurrentValue:   position.Value,
			TotalSpent:     totalSpent,
			Profit:         profit,
			Position:       positionType,
			YesSharesOwned: position.YesSharesOwned,
			NoSharesOwned:  position.NoSharesOwned,
			EarliestBet:    earliestBet,
		})
	}

	// Sort by profit (descending), then by earliest bet time (ascending) for ties
	sort.Slice(leaderboard, func(i, j int) bool {
		if leaderboard[i].Profit == leaderboard[j].Profit {
			// If profits are equal, rank by who bet earlier (ascending time)
			return leaderboard[i].EarliestBet.Before(leaderboard[j].EarliestBet)
		}
		// Otherwise rank by profit (descending)
		return leaderboard[i].Profit > leaderboard[j].Profit
	})

	// Assign ranks
	for i := range leaderboard {
		leaderboard[i].Rank = i + 1
	}

	return leaderboard, nil
}

// GlobalUserProfitability represents a user's total profitability across all markets
type GlobalUserProfitability struct {
	Username          string    `json:"username"`
	TotalProfit       int64     `json:"totalProfit"`
	TotalCurrentValue int64     `json:"totalCurrentValue"`
	TotalSpent        int64     `json:"totalSpent"`
	ActiveMarkets     int       `json:"activeMarkets"`   // Number of markets with positions
	ResolvedMarkets   int       `json:"resolvedMarkets"` // Number of resolved markets participated
	EarliestBet       time.Time `json:"earliestBet"`
	Rank              int       `json:"rank"`
}

// CalculateGlobalLeaderboard calculates profitability rankings for all users across all markets
func CalculateGlobalLeaderboard(db *gorm.DB) ([]GlobalUserProfitability, error) {
	if db == nil {
		return nil, errors.New("Failed to fetch users from database: database connection is nil")
	}

	// Get all users who have made bets
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		ErrorLogger(err, "Failed to fetch users from database.")
		return nil, err
	}

	if len(users) == 0 {
		return []GlobalUserProfitability{}, nil
	}

	var globalLeaderboard []GlobalUserProfitability

	for _, user := range users {
		// Get all market positions for this user
		userPositions, err := CalculateAllUserMarketPositions_WPAM_DBPM(db, user.Username)
		if err != nil {
			ErrorLogger(err, "Failed to calculate user positions for "+user.Username)
			continue // Skip this user but continue with others
		}

		// Skip users with no positions
		if len(userPositions) == 0 {
			continue
		}

		var totalProfit int64 = 0
		var totalCurrentValue int64 = 0
		var totalSpent int64 = 0
		var activeMarkets int = 0
		var resolvedMarkets int = 0
		var earliestBet time.Time
		var hasEarliestBet bool = false

		// Get all bets for this user to find earliest bet time
		var userBets []models.Bet
		if err := db.Where("username = ?", user.Username).Order("placed_at ASC").Find(&userBets).Error; err != nil {
			ErrorLogger(err, "Failed to fetch bets for user "+user.Username)
			continue
		}

		if len(userBets) > 0 {
			earliestBet = userBets[0].PlacedAt
			hasEarliestBet = true
		}

		// Aggregate profits from all markets
		for _, position := range userPositions {
			// Calculate profit for this market: currentValue - totalSpent
			marketProfit := position.Value - position.TotalSpent

			totalProfit += marketProfit
			totalCurrentValue += position.Value
			totalSpent += position.TotalSpent

			// Count market types
			if position.IsResolved {
				resolvedMarkets++
			} else {
				activeMarkets++
			}
		}

		// Only include users with some betting activity
		if hasEarliestBet {
			globalLeaderboard = append(globalLeaderboard, GlobalUserProfitability{
				Username:          user.Username,
				TotalProfit:       totalProfit,
				TotalCurrentValue: totalCurrentValue,
				TotalSpent:        totalSpent,
				ActiveMarkets:     activeMarkets,
				ResolvedMarkets:   resolvedMarkets,
				EarliestBet:       earliestBet,
			})
		}
	}

	// Sort by total profit (descending), then by earliest bet time (ascending) for ties
	sort.Slice(globalLeaderboard, func(i, j int) bool {
		if globalLeaderboard[i].TotalProfit == globalLeaderboard[j].TotalProfit {
			// If profits are equal, rank by who bet earlier (ascending time)
			return globalLeaderboard[i].EarliestBet.Before(globalLeaderboard[j].EarliestBet)
		}
		// Otherwise rank by total profit (descending)
		return globalLeaderboard[i].TotalProfit > globalLeaderboard[j].TotalProfit
	})

	// Assign ranks
	for i := range globalLeaderboard {
		globalLeaderboard[i].Rank = i + 1
	}

	return globalLeaderboard, nil
}
