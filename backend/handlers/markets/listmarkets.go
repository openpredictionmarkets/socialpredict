package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"socialpredict/models"
	"socialpredict/util"
)

func ListMarketsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	db := util.GetDB()

	var countMarketTotal int64
	countResult := db.Raw("SELECT COUNT(*) FROM markets;").Scan(&countMarketTotal)
	if countResult.Error != nil {
		// Handle error
		log.Printf("Error executing raw SQL query: %v", countResult.Error)
	} else {
		log.Printf("Total number of markets: %d", countMarketTotal)
	}

	var markets []models.Market
	result := db.Order("RANDOM()").Limit(int(countMarketTotal)).Find(&markets) // Adjust this line based on your ORM
	if result.Error != nil {
		http.Error(w, "Error fetching markets", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header
	w.Header().Set("Content-Type", "application/json")

	// Encode the markets slice to JSON and send it as a response
	if err := json.NewEncoder(w).Encode(markets); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
