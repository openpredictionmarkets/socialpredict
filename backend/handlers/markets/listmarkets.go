package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"socialpredict/models"
	"socialpredict/util"

	"gorm.io/gorm"
)

// ListMarketsHandler handles the HTTP request for listing markets.
func ListMarketsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	db := util.GetDB()
	markets, err := ListMarkets(db)
	if err != nil {
		http.Error(w, "Error fetching markets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(markets); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ListMarkets fetches a random list of all markets from the database.
func ListMarkets(db *gorm.DB) ([]models.Market, error) {
	var markets []models.Market
	result := db.Order("RANDOM()").Limit(100).Find(&markets) // Set a reasonable limit
	if result.Error != nil {
		log.Printf("Error fetching markets: %v", result.Error)
		return nil, result.Error
	}

	return markets, nil
}
