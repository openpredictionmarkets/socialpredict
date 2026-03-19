package publicuser

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"socialpredict/models"
	"socialpredict/util"
)

// BetHistoryItem is a single bet enriched with the market question title.
type BetHistoryItem struct {
	ID            uint      `json:"id"`
	MarketID      uint      `json:"marketId"`
	QuestionTitle string    `json:"questionTitle"`
	Action        string    `json:"action"`  // "BUY" or "SELL"
	Outcome       string    `json:"outcome"` // "YES" or "NO"
	Amount        int64     `json:"amount"`
	PlacedAt      time.Time `json:"placedAt"`
}

// GetBetHistory returns all individual bets placed by a user, newest first,
// enriched with the market question title.
func GetBetHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	db := util.GetDB()

	bets, err := fetchUserBets(db, username)
	if err != nil {
		log.Printf("Error fetching bet history for %s: %v", username, err)
		http.Error(w, "Error fetching bet history", http.StatusInternalServerError)
		return
	}

	items, err := enrichBetsWithMarketTitles(db, bets)
	if err != nil {
		log.Printf("Error enriching bet history: %v", err)
		http.Error(w, "Error enriching bet history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// enrichBetsWithMarketTitles joins bets with their market's question title.
func enrichBetsWithMarketTitles(db *gorm.DB, bets []models.Bet) ([]BetHistoryItem, error) {
	// Pre-fetch all distinct markets to avoid N+1 queries.
	marketIDs := make([]uint, 0, len(bets))
	seen := make(map[uint]bool)
	for _, b := range bets {
		if !seen[b.MarketID] {
			seen[b.MarketID] = true
			marketIDs = append(marketIDs, b.MarketID)
		}
	}

	var markets []models.Market
	if err := db.Where("id IN ?", marketIDs).Find(&markets).Error; err != nil {
		return nil, err
	}

	titleByID := make(map[uint]string, len(markets))
	for _, m := range markets {
		titleByID[uint(m.ID)] = m.QuestionTitle
	}

	items := make([]BetHistoryItem, len(bets))
	for i, b := range bets {
		action := "BUY"
		if b.Amount < 0 {
			action = "SELL"
		}
		items[i] = BetHistoryItem{
			ID:            b.ID,
			MarketID:      b.MarketID,
			QuestionTitle: titleByID[b.MarketID],
			Action:        action,
			Outcome:       b.Outcome,
			Amount:        b.Amount,
			PlacedAt:      b.PlacedAt,
		}
	}
	return items, nil
}
