package buybetshandlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"socialpredict/handlers/bets/dto"
	dbets "socialpredict/internal/domain/bets"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

// PlaceBetHandler returns an HTTP handler that delegates bet placement to the bets domain service.
func PlaceBetHandler(betsSvc dbets.ServiceInterface, usersSvc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		user, httpErr := authsvc.ValidateUserAndEnforcePasswordChangeGetUser(r, usersSvc)
		if httpErr != nil {
			http.Error(w, httpErr.Error(), httpErr.StatusCode)
			return
		}

		req, decodeErr := decodePlaceBetRequest(r)
		if decodeErr != nil {
			http.Error(w, decodeErr.Error(), http.StatusBadRequest)
			return
		}

		placedBet, err := betsSvc.Place(r.Context(), toPlaceRequest(req, user.Username))
		if err != nil {
			writePlaceBetError(w, err)
			return
		}

		writePlaceBetResponse(w, placedBet)
	}
}

func decodePlaceBetRequest(r *http.Request) (dto.PlaceBetRequest, error) {
	var req dto.PlaceBetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return dto.PlaceBetRequest{}, errors.New("Invalid request body")
	}
	return req, nil
}

func toPlaceRequest(req dto.PlaceBetRequest, username string) dbets.PlaceRequest {
	return dbets.PlaceRequest{
		Username: username,
		MarketID: req.MarketID,
		Amount:   req.Amount,
		Outcome:  req.Outcome,
	}
}

func writePlaceBetError(w http.ResponseWriter, err error) {
	switch err {
	case dbets.ErrInvalidOutcome, dbets.ErrInvalidAmount:
		http.Error(w, err.Error(), http.StatusBadRequest)
	case dbets.ErrMarketClosed:
		http.Error(w, err.Error(), http.StatusConflict)
	case dbets.ErrInsufficientBalance:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	case dmarkets.ErrMarketNotFound:
		http.Error(w, "Market not found", http.StatusNotFound)
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func writePlaceBetResponse(w http.ResponseWriter, placedBet *dbets.PlacedBet) {
	response := dto.PlaceBetResponse{
		Username: placedBet.Username,
		MarketID: placedBet.MarketID,
		Amount:   placedBet.Amount,
		Outcome:  placedBet.Outcome,
		PlacedAt: placedBet.PlacedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}
