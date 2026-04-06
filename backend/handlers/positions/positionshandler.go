package positions

import (
	"encoding/json"
	"errors"
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

		marketID, err := parseMarketID(mux.Vars(r)["marketId"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		positions, err := svc.GetMarketPositions(r.Context(), marketID)
		if err != nil {
			writePositionsError(w, marketID, err)
			return
		}

		responses := mapPositionsToResponses(positions)

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

		marketID, username, err := parseMarketUserParams(mux.Vars(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		position, err := svc.GetUserPositionInMarket(r.Context(), marketID, username)
		if err != nil {
			writeUserPositionError(w, marketID, username, err)
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

func writePositionsError(w http.ResponseWriter, marketID int64, err error) {
	switch err {
	case dmarkets.ErrMarketNotFound:
		http.Error(w, "Market not found", http.StatusNotFound)
	case dmarkets.ErrInvalidInput:
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
	default:
		log.Printf("Error getting market positions for market %d: %v", marketID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func mapPositionsToResponses(positions []*dmarkets.UserPosition) []userPositionResponse {
	responses := make([]userPositionResponse, 0, len(positions))
	for _, pos := range positions {
		if pos == nil {
			continue
		}
		responses = append(responses, newUserPositionResponse(pos))
	}
	return responses
}

func parseMarketUserParams(vars map[string]string) (int64, string, error) {
	marketID, err := parseMarketID(vars["marketId"])
	if err != nil {
		return 0, "", err
	}

	username := vars["username"]
	if username == "" {
		return 0, "", errors.New("Username is required")
	}

	return marketID, username, nil
}

func writeUserPositionError(w http.ResponseWriter, marketID int64, username string, err error) {
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
}
