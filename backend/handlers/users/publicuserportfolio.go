package usershandlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"socialpredict/models"
	"socialpredict/util"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type PortfolioItem struct {
	MarketID      uint      `json:"marketId"`
	QuestionTitle string    `json:"questionTitle"`
	TotalYesBets  int64     `json:"totalYesBets"`
	TotalNoBets   int64     `json:"totalNoBets"`
	LastBetPlaced time.Time `json:"lastBetPlaced"`
}

func GetPublicUserPortfolio(w http.ResponseWriter, r *http.Request) {
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

	// fetch all markets that user has bet on, order by most recently bet on
	userportfolio, err := processUserBets(userbets)
	if err != nil {
		log.Printf("Error processing user bets: %v", err)
		http.Error(w, "Error processing user bets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userportfolio)
}

// fetchUserBets retrieves all bets made by a specific user
func fetchUserBets(db *gorm.DB, username string) ([]models.Bet, error) {
	// First, get the user info
	userInfo := GetPublicUserInfo(db, username)
	if userInfo.Username == "" {
		log.Printf("User not found for username: %v", username)
		return nil, errors.New("user not found")
	}

	var userbets []models.Bet
	// Retrieve all bets made by the user
	if err := db.Where("username = ?", userInfo.Username).Order("placed_at desc").Find(&userbets).Error; err != nil {
		return nil, err
	}

	return userbets, nil
}

func processUserBets(userbets []models.Bet) ([]PortfolioItem, error) {
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

		// Aggregate YES and NO bets
		if bet.Outcome == "YES" {
			item.TotalYesBets += bet.Amount
		} else if bet.Outcome == "NO" {
			item.TotalNoBets += bet.Amount
		}

		// Update the last bet placed time if this bet is more recent
		if bet.PlacedAt.After(item.LastBetPlaced) {
			item.LastBetPlaced = bet.PlacedAt
		}

		// Put the item back in the map
		marketMap[bet.MarketID] = item
	}

	db := util.GetDB()

	// Fetch and append market names
	for marketID, item := range marketMap {
		var market models.Market
		if err := db.Where("id = ?", marketID).First(&market).Error; err != nil {
			return nil, err
		}
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
