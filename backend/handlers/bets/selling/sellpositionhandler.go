package sellbetshandlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"socialpredict/handlers/bets/dto"
	bets "socialpredict/internal/domain/bets"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

// SellPositionHandler returns an HTTP handler that delegates sales to the bets service.
func SellPositionHandler(betsSvc bets.ServiceInterface, usersSvc dusers.ServiceInterface) http.HandlerFunc {
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

		req, decodeErr := decodeSellRequest(r)
		if decodeErr != nil {
			http.Error(w, decodeErr.Error(), http.StatusBadRequest)
			return
		}

		result, err := betsSvc.Sell(r.Context(), toSellRequest(req, user.Username))
		if err != nil {
			handleSellError(w, err)
			return
		}

		writeSellResponse(w, result)
	}
}

func decodeSellRequest(r *http.Request) (dto.SellBetRequest, error) {
	var req dto.SellBetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return dto.SellBetRequest{}, errors.New("Invalid request body")
	}
	return req, nil
}

func toSellRequest(req dto.SellBetRequest, username string) bets.SellRequest {
	return bets.SellRequest{
		Username: username,
		MarketID: req.MarketID,
		Amount:   req.Amount,
		Outcome:  req.Outcome,
	}
}

func handleSellError(w http.ResponseWriter, err error) {
	if dustErr, ok := err.(bets.ErrDustCapExceeded); ok {
		http.Error(w, dustErr.Error(), http.StatusUnprocessableEntity)
		return
	}

	switch {
	case errors.Is(err, bets.ErrInvalidOutcome), errors.Is(err, bets.ErrInvalidAmount):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, bets.ErrMarketClosed):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, bets.ErrNoPosition), errors.Is(err, bets.ErrInsufficientShares):
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	case errors.Is(err, dmarkets.ErrMarketNotFound):
		http.Error(w, "Market not found", http.StatusNotFound)
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func writeSellResponse(w http.ResponseWriter, result *bets.SellResult) {
	response := dto.SellBetResponse{
		Username:      result.Username,
		MarketID:      result.MarketID,
		SharesSold:    result.SharesSold,
		SaleValue:     result.SaleValue,
		Dust:          result.Dust,
		Outcome:       result.Outcome,
		TransactionAt: result.TransactionAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}
