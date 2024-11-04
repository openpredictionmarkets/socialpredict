package publicuser

import (
	"encoding/json"
	"log"
	"net/http"
	"socialpredict/handlers/positions"
	"socialpredict/models"
	"socialpredict/util"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type PortfolioItem struct {
	MarketID       uint      `json:"marketId"`
	QuestionTitle  string    `json:"questionTitle"`
	YesSharesOwned int64     `json:"yesSharesOwned"`
	NoSharesOwned  int64     `json:"noSharesOwned"`
	LastBetPlaced  time.Time `json:"lastBetPlaced"`
}

type PortfolioTotal struct {
	PortfolioItems   []PortfolioItem `json:"portfolioItems"`
	TotalSharesOwned int64           `json:"totalSharesOwned"`
}

func GetPortfolio(w http.ResponseWriter, r *http.Request) {
	// Extract the username from the URL
	vars := mux.Vars(r)
	username := vars["username"]

	db := util.GetDB()

	// fetch all bets made by a specific user
	userbets, err := fetchUserBets(db, username)
	if err != nil {
		log.Printf("Error fetching user bets: %v", err)
		http.Error(w, "Error fetching user bets", http.StatusInternalServerError)
		return
	}

	// Create a market map from the user's bets
	marketMap := makeUserMarketMap(userbets)

	// Process the market map to calculate positions and fetch market titles
	userPositionsPortfolio, err := processMarketMap(db, marketMap, username)
	if err != nil {
		log.Printf("Error processing market map: %v", err)
		http.Error(w, "Error processing market map", http.StatusInternalServerError)
		return
	}

	totalSharesOwned := calculateTotalShares(userPositionsPortfolio)

	portfolioTotal := PortfolioTotal{
		PortfolioItems:   userPositionsPortfolio,
		TotalSharesOwned: totalSharesOwned,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(portfolioTotal)
}

// fetchUserBets retrieves all bets made by a specific user
func fetchUserBets(db *gorm.DB, username string) ([]models.Bet, error) {
	var userbets []models.Bet
	// Retrieve all bets made by the user
	if err := db.Where("username = ?", username).Order("placed_at desc").Find(&userbets).Error; err != nil {
		return nil, err
	}

	return userbets, nil
}

// makeUserMarketMap creates a map of PortfolioItem from the user's bets
func makeUserMarketMap(userbets []models.Bet) map[uint]PortfolioItem {
	marketMap := make(map[uint]PortfolioItem)

	// Iterate over all bets
	for _, bet := range userbets {
		// Check if this market is already in our map
		item, exists := marketMap[bet.MarketID]
		if !exists {
			item = PortfolioItem{
				MarketID:      bet.MarketID,
				LastBetPlaced: bet.PlacedAt,
			}
		}

		// Update the last bet placed time if this bet is more recent
		if bet.PlacedAt.After(item.LastBetPlaced) {
			item.LastBetPlaced = bet.PlacedAt
		}

		// Put the item back in the map
		marketMap[bet.MarketID] = item
	}

	return marketMap
}

func processMarketMap(db *gorm.DB, marketMap map[uint]PortfolioItem, username string) ([]PortfolioItem, error) {
	// Calculate market positions for each market
	for marketID := range marketMap {
		position, err := positions.CalculateMarketPositionForUser_WPAM_DBPM(db, strconv.Itoa(int(marketID)), username)
		if err != nil {
			return nil, err
		}

		// Fetch market title
		var market models.Market
		if err := db.Where("id = ?", marketID).First(&market).Error; err != nil {
			return nil, err
		}

		// Update the market item with the calculated positions and market title
		item := marketMap[marketID]
		item.YesSharesOwned = position.YesSharesOwned
		item.NoSharesOwned = position.NoSharesOwned
		item.QuestionTitle = market.QuestionTitle
		marketMap[marketID] = item
	}

	// Convert map to slice
	var userportfolio []PortfolioItem
	for _, item := range marketMap {
		userportfolio = append(userportfolio, item)
	}

	// Sort the portfolio by LastBetPlaced in descending order
	sort.Slice(userportfolio, func(i, j int) bool {
		return userportfolio[i].LastBetPlaced.After(userportfolio[j].LastBetPlaced)
	})

	return userportfolio, nil
}

func calculateTotalShares(portfolio []PortfolioItem) int64 {
	var totalShares int64
	for _, item := range portfolio {
		totalShares += item.YesSharesOwned + item.NoSharesOwned
	}
	return totalShares
}
