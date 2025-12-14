package betshandlers

import (
	"encoding/json"
	"errors"
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

		marketID, err := parseMarketID(mux.Vars(r)["marketId"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		betsDisplayInfo, err := svc.GetMarketBets(r.Context(), marketID)
		if err != nil {
			writeMarketBetsError(w, marketID, err)
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

func parseMarketID(marketIDStr string) (int64, error) {
	if marketIDStr == "" {
		return 0, errors.New("Market ID is required")
	}

	marketID, err := strconv.ParseInt(marketIDStr, 10, 64)
	if err != nil {
		return 0, errors.New("Invalid market ID")
	}
	return marketID, nil
}

func writeMarketBetsError(w http.ResponseWriter, marketID int64, err error) {
	switch err {
	case dmarkets.ErrMarketNotFound:
		http.Error(w, "Market not found", http.StatusNotFound)
	case dmarkets.ErrInvalidInput:
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
	default:
		log.Printf("Error getting market bets for market %d: %v", marketID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
