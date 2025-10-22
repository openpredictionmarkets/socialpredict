package marketshandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	dmarkets "socialpredict/internal/domain/markets"

	"github.com/gorilla/mux"
)

// MarketBetsHandlerWithService creates a service-injected bets handler
func MarketBetsHandlerWithService(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		// Parse market ID from URL
		vars := mux.Vars(r)
		marketIdStr := vars["marketId"]
		if marketIdStr == "" {
			http.Error(w, "Market ID is required", http.StatusBadRequest)
			return
		}

		// Convert marketId to int64
		marketID, err := strconv.ParseInt(marketIdStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid market ID", http.StatusBadRequest)
			return
		}

		// Call domain service
		betsDisplayInfo, err := svc.GetMarketBets(r.Context(), marketID)
		if err != nil {
			// Map domain errors to HTTP status codes
			switch err {
			case dmarkets.ErrMarketNotFound:
				http.Error(w, "Market not found", http.StatusNotFound)
			case dmarkets.ErrInvalidInput:
				http.Error(w, "Invalid market ID", http.StatusBadRequest)
			default:
				log.Printf("Error getting market bets for market %d: %v", marketID, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Respond with the bets display information
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(betsDisplayInfo); err != nil {
			log.Printf("Error encoding bets response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	}
}
