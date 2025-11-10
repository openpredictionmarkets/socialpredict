package positions

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	dmarkets "socialpredict/internal/domain/markets"

	"github.com/gorilla/mux"
)

// MarketPositionsHandlerWithService creates a service-injected positions handler for all users
func MarketPositionsHandlerWithService(svc dmarkets.ServiceInterface) http.HandlerFunc {
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
		positions, err := svc.GetMarketPositions(r.Context(), marketID)
		if err != nil {
			// Map domain errors to HTTP status codes
			switch err {
			case dmarkets.ErrMarketNotFound:
				http.Error(w, "Market not found", http.StatusNotFound)
			case dmarkets.ErrInvalidInput:
				http.Error(w, "Invalid market ID", http.StatusBadRequest)
			default:
				log.Printf("Error getting market positions for market %d: %v", marketID, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Map to DTO responses
		responses := make([]userPositionResponse, 0, len(positions))
		for _, pos := range positions {
			if pos == nil {
				continue
			}
			responses = append(responses, newUserPositionResponse(pos))
		}

		// Respond with the positions information
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(responses); err != nil {
			log.Printf("Error encoding positions response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	}
}

// MarketUserPositionHandlerWithService creates a service-injected handler for a specific user's position
func MarketUserPositionHandlerWithService(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		// Parse market ID and username from URL
		vars := mux.Vars(r)
		marketIdStr := vars["marketId"]
		username := vars["username"]

		if marketIdStr == "" {
			http.Error(w, "Market ID is required", http.StatusBadRequest)
			return
		}
		if username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}

		// Convert marketId to int64
		marketID, err := strconv.ParseInt(marketIdStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid market ID", http.StatusBadRequest)
			return
		}

		// Call domain service
		position, err := svc.GetUserPositionInMarket(r.Context(), marketID, username)
		if err != nil {
			// Map domain errors to HTTP status codes
			switch err {
			case dmarkets.ErrMarketNotFound:
				http.Error(w, "Market not found", http.StatusNotFound)
			case dmarkets.ErrUserNotFound:
				http.Error(w, "User not found", http.StatusNotFound)
			case dmarkets.ErrInvalidInput:
				http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			default:
				log.Printf("Error getting user position for market %d, user %s: %v", marketID, username, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Respond with the user position information
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(newUserPositionResponse(position)); err != nil {
			log.Printf("Error encoding user position response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	}
}
